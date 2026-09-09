// Тестовый стенд. Менять здесь нечего: проверяем этими же тестами.
//
// Собирает ./cmd/server, запускает на свободном порту через PORT, ждёт первого
// соединения и дальше стучится по обычному HTTP. Ваш код стенд не импортирует,
// видит ровно то же, что увидел бы браузер.
package tests

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var baseURL string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	moduleRoot, err := filepath.Abs("..")
	if err != nil {
		fmt.Fprintln(os.Stderr, "стенд: не нашёл корень модуля:", err)
		return 1
	}

	tmp, err := os.MkdirTemp("", "entrytest-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "стенд: не создал временный каталог:", err)
		return 1
	}
	defer os.RemoveAll(tmp)

	bin := filepath.Join(tmp, "server")
	build := exec.Command("go", "build", "-o", bin, "./cmd/server")
	build.Dir = moduleRoot
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "\nСервер не компилируется, начните с этого:\n\n%s\n", out)
		return 1
	}

	port, err := freePort()
	if err != nil {
		fmt.Fprintln(os.Stderr, "стенд: не нашёл свободный порт:", err)
		return 1
	}
	baseURL = "http://127.0.0.1:" + port

	var serverLog bytes.Buffer
	server := exec.Command(bin)
	server.Dir = moduleRoot // чтобы "frontend" искался там же, где при `go run ./cmd/server`
	server.Env = append(os.Environ(), "PORT="+port)
	server.Stdout = &serverLog
	server.Stderr = &serverLog
	if err := server.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "стенд: не запустил ваш сервер:", err)
		return 1
	}
	defer server.Process.Kill()

	if !waitReachable("127.0.0.1:"+port, 10*time.Second) {
		fmt.Fprintf(os.Stderr, "\nСервер за 10 секунд так и не начал слушать PORT=%s.\n"+
			"Проверьте, что порт берётся из окружения. Что он успел написать:\n\n%s\n",
			port, serverLog.String())
		return 1
	}

	code := m.Run()
	if code != 0 && serverLog.Len() > 0 {
		fmt.Fprintf(os.Stderr, "\n--- логи вашего сервера ---\n%s\n", serverLog.String())
	}
	return code
}

func freePort() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	_, port, err := net.SplitHostPort(l.Addr().String())
	return port, err
}

func waitReachable(addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 250*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// таймаут, чтобы зависший обработчик падал быстро и не держал весь прогон.
var client = &http.Client{Timeout: 5 * time.Second}
