package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// ---------- Этап 1: здоровье + фронтенд ----------

func TestStage1_Health(t *testing.T) {
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Этап 1: GET /health не дошёл до сервера: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Этап 1: GET /health вернул %d, ожидается 200", resp.StatusCode)
	}
	if got := strings.TrimSpace(string(body)); got != "ok" {
		t.Fatalf("Этап 1: GET /health вернул тело %q, ожидается %q", got, "ok")
	}
}

func TestStage1_Frontend(t *testing.T) {
	resp, err := client.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("Этап 1: GET / не дошёл до сервера: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Этап 1: GET / вернул %d, ожидается 200 со статикой из frontend/", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("Этап 1: GET / вернул Content-Type %q, ожидается text/html", ct)
	}
	if !bytes.Contains(bytes.ToLower(body), []byte("<html")) {
		t.Fatalf("Этап 1: GET / вернул что-то другое, не frontend/index.html")
	}
}

// ---------- Этап 2: эхо обычного текста ----------

func TestStage2_EchoText(t *testing.T) {
	for _, payload := range []string{
		"hello from C++ land",
		"line one\nline two",
		`{"this": "is NOT json because the Content-Type is text/plain"}`,
	} {
		resp, err := client.Post(baseURL+"/echo", "text/plain", strings.NewReader(payload))
		if err != nil {
			t.Fatalf("Этап 2: POST /echo не дошёл до сервера: %v", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Этап 2: POST /echo вернул %d, ожидается 200", resp.StatusCode)
		}
		if string(body) != payload {
			t.Fatalf("Этап 2: POST /echo: отправлено %q, вернулось %q, ожидаются те же байты", payload, string(body))
		}
	}
}

// ---------- Этап 3: JSON-эхо ----------

func TestStage3_EchoJSON(t *testing.T) {
	in := map[string]string{"message": "hello, JSON"}
	resp := postJSON(t, "/echo", in)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Этап 3: POST /echo с Content-Type application/json вернул %d, ожидается 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("Этап 3: POST /echo (JSON) вернул Content-Type %q, ожидается application/json", ct)
	}
	var out struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("Этап 3: POST /echo (JSON) вернул не JSON: %v", err)
	}
	if out.Message != in["message"] {
		t.Fatalf("Этап 3: POST /echo (JSON): отправлено %q, вернулось %q", in["message"], out.Message)
	}
}

func TestStage3_EchoJSONMalformed(t *testing.T) {
	resp, err := client.Post(baseURL+"/echo", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatalf("Этап 3: POST /echo не дошёл до сервера: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Этап 3: POST /echo со сломанным JSON вернул %d, ожидается 400", resp.StatusCode)
	}
}

// ---------- Этап 4: создание сообщений ----------

func TestStage4_CreateMessage(t *testing.T) {
	resp := postJSON(t, "/messages", map[string]string{"message": "first message"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Этап 4: POST /messages вернул %d, ожидается 201", resp.StatusCode)
	}
	m := decodeMessage(t, resp.Body, "Этап 4: ответ POST /messages")
	if m.Message != "first message" {
		t.Fatalf("Этап 4: POST /messages вернул message %q, ожидается %q", m.Message, "first message")
	}
	if m.CreatedAt == "" {
		t.Fatalf("Этап 4: в ответе POST /messages нет created_at")
	}
	if _, err := time.Parse(time.RFC3339, m.CreatedAt); err != nil {
		t.Fatalf("Этап 4: created_at %q не разбирается как RFC 3339", m.CreatedAt)
	}
}

func TestStage4_UniqueIDs(t *testing.T) {
	seen := map[int64]bool{}
	for i := 0; i < 3; i++ {
		resp := postJSON(t, "/messages", map[string]string{"message": fmt.Sprintf("unique-id-check %d", i)})
		m := decodeMessage(t, resp.Body, "Этап 4: ответ POST /messages")
		resp.Body.Close()
		if seen[m.ID] {
			t.Fatalf("Этап 4: POST /messages выдал id %d второй раз, id должен быть свой у каждого сообщения", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestStage4_RejectsBadInput(t *testing.T) {
	for name, body := range map[string]string{
		"со сломанным JSON":     "{not json",
		"с пустым сообщением":   `{"message": ""}`,
		"с отсутствующим полем": `{}`,
	} {
		resp, err := client.Post(baseURL+"/messages", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("Этап 4: POST /messages не дошёл до сервера: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Этап 4: POST /messages %s вернул %d, ожидается 400", name, resp.StatusCode)
		}
	}
}

// ---------- Этап 5: список, новые сверху ----------

func TestStage5_ListNewestFirst(t *testing.T) {
	posted := make([]apiMessage, 0, 3)
	for i := 0; i < 3; i++ {
		resp := postJSON(t, "/messages", map[string]string{"message": fmt.Sprintf("ordering-check %d", i)})
		posted = append(posted, decodeMessage(t, resp.Body, "Этап 5: ответ POST /messages"))
		resp.Body.Close()
	}

	resp, err := client.Get(baseURL + "/messages")
	if err != nil {
		t.Fatalf("Этап 5: GET /messages не дошёл до сервера: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Этап 5: GET /messages вернул %d, ожидается 200", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	if strings.TrimSpace(string(raw)) == "null" {
		t.Fatalf("Этап 5: GET /messages вернул null, пустой список должен быть []")
	}
	var list []apiMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatalf("Этап 5: GET /messages вернул не массив сообщений, а: %.200s", raw)
	}
	if len(list) < 3 {
		t.Fatalf("Этап 5: отправлено минимум 3 сообщения, в списке %d", len(list))
	}
	// Три только что отправленных должны быть первыми тремя, в обратном порядке отправки.
	for i := 0; i < 3; i++ {
		want := posted[2-i]
		got := list[i]
		if got.ID != want.ID || got.Message != want.Message {
			t.Fatalf("Этап 5: на позиции %d стоит %q (id %d), а новые идут сверху, значит ожидается %q (id %d)",
				i, got.Message, got.ID, want.Message, want.ID)
		}
	}
}

// ---------- Этап 6: удаление ----------

func TestStage6_Delete(t *testing.T) {
	resp := postJSON(t, "/messages", map[string]string{"message": "delete me"})
	m := decodeMessage(t, resp.Body, "Этап 6: ответ POST /messages")
	resp.Body.Close()

	if status := doDelete(t, m.ID); status != http.StatusNoContent {
		t.Fatalf("Этап 6: DELETE /messages/%d вернул %d, ожидается 204", m.ID, status)
	}

	listResp, err := client.Get(baseURL + "/messages")
	if err != nil {
		t.Fatalf("Этап 6: GET /messages не дошёл до сервера: %v", err)
	}
	var list []apiMessage
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatalf("Этап 6: список из GET /messages не разбирается: %v", err)
	}
	listResp.Body.Close()
	for _, got := range list {
		if got.ID == m.ID {
			t.Fatalf("Этап 6: сообщение с id %d удалено, но всё ещё в GET /messages", m.ID)
		}
	}

	if status := doDelete(t, m.ID); status != http.StatusNotFound {
		t.Fatalf("Этап 6: повторный DELETE /messages/%d вернул %d, ожидается 404", m.ID, status)
	}
}

func TestStage6_DeleteUnknownID(t *testing.T) {
	if status := doDelete(t, 999999999); status != http.StatusNotFound {
		t.Fatalf("Этап 6: DELETE /messages/999999999 (такого id не было) вернул %d, ожидается 404", status)
	}
}

// ---------- вспомогательное ----------

type apiMessage struct {
	ID        int64  `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func postJSON(t *testing.T, path string, payload any) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("стенд: не сериализовал данные: %v", err)
	}
	resp, err := client.Post(baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s не дошёл до сервера: %v", path, err)
	}
	return resp
}

func decodeMessage(t *testing.T, r io.Reader, what string) apiMessage {
	t.Helper()
	raw, _ := io.ReadAll(r)
	var m apiMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf(`%s: ожидается JSON вида {"id": 1, "message": "...", "created_at": "..."}, пришло: %.200s`, what, raw)
	}
	return m
}

func doDelete(t *testing.T, id int64) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/messages/%d", baseURL, id), nil)
	if err != nil {
		t.Fatalf("стенд: не собрал DELETE-запрос: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /messages/%d не дошёл до сервера: %v", id, err)
	}
	resp.Body.Close()
	return resp.StatusCode
}
