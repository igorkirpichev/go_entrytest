/* Фронтенд вступительного теста. Трогать не нужно: он просто ходит
   в HTTP-API, описанный в README. */

const $ = (sel) => document.querySelector(sel);

/* ---------- чек-лист этапов ---------- */

const stageChecks = {
  1: async () => {
    const r = await fetch("/health");
    return r.ok && (await r.text()).trim() === "ok";
  },
  2: async () => {
    const probe = "stage-2-probe";
    const r = await fetch("/echo", {
      method: "POST",
      headers: { "Content-Type": "text/plain" },
      body: probe,
    });
    return r.ok && (await r.text()) === probe;
  },
  3: async () => {
    const r = await fetch("/echo", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: "stage-3-probe" }),
    });
    if (!r.ok) return false;
    const data = await r.json();
    return data.message === "stage-3-probe";
  },
  // Этапы 4–6 проверяются вместе: создать сообщение, найти его в списке, удалить.
  4: async (ctx) => {
    const r = await fetch("/messages", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: "проверка этапа" }),
    });
    if (r.status !== 201) return false;
    const m = await r.json();
    ctx.probeId = m.id;
    return m.id !== undefined && m.message && m.created_at;
  },
  5: async (ctx) => {
    const r = await fetch("/messages");
    if (!r.ok) return false;
    const list = await r.json();
    return Array.isArray(list) && list.some((m) => m.id === ctx.probeId);
  },
  6: async (ctx) => {
    if (ctx.probeId === undefined) return false;
    const r = await fetch(`/messages/${ctx.probeId}`, { method: "DELETE" });
    return r.status === 204;
  },
};

async function runStageChecks() {
  const ctx = {};
  for (const [stage, check] of Object.entries(stageChecks)) {
    const li = $(`#stages li[data-stage="${stage}"]`);
    let pass = false;
    try {
      pass = await check(ctx);
    } catch {
      pass = false;
    }
    li.classList.toggle("pass", pass);
    li.classList.toggle("fail", !pass);
  }
  refreshMessages();
}

/* ---------- индикатор /health ---------- */

async function refreshHealth() {
  const el = $("#health");
  try {
    const r = await fetch("/health");
    const ok = r.ok && (await r.text()).trim() === "ok";
    el.className = "health " + (ok ? "ok" : "bad");
    el.textContent = ok ? "● сервер отвечает" : "● /health отвечает не ok";
  } catch {
    el.className = "health bad";
    el.textContent = "● сервер не отвечает";
  }
}

/* ---------- эхо ---------- */

$("#echo-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const out = $("#echo-output");
  const text = $("#echo-input").value;
  const asJSON = $("#echo-json").checked;
  try {
    const r = await fetch("/echo", {
      method: "POST",
      headers: { "Content-Type": asJSON ? "application/json" : "text/plain" },
      body: asJSON ? JSON.stringify({ message: text }) : text,
    });
    const body = await r.text();
    out.className = "output " + (r.ok ? "ok" : "bad");
    out.textContent = `HTTP ${r.status}\n${body}`;
  } catch (err) {
    out.className = "output bad";
    out.textContent = "запрос не дошёл: " + err;
  }
});

/* ---------- доска сообщений ---------- */

async function refreshMessages() {
  const status = $("#board-status");
  const ul = $("#messages");
  try {
    const r = await fetch("/messages");
    if (!r.ok) {
      status.textContent = `GET /messages вернул ${r.status}, доска заработает после этапа 5.`;
      return;
    }
    const list = await r.json();
    if (!Array.isArray(list)) {
      status.textContent = "GET /messages вернул не JSON-массив, возможно null.";
      return;
    }
    status.textContent = list.length ? "" : "Сообщений нет.";
    ul.replaceChildren(
      ...list.map((m) => {
        const li = document.createElement("li");
        const when = document.createElement("span");
        when.className = "when";
        when.textContent = m.created_at ? new Date(m.created_at).toLocaleTimeString() : "";
        const text = document.createElement("span");
        text.className = "text";
        text.textContent = m.message;
        const del = document.createElement("button");
        del.className = "delete";
        del.textContent = "✕";
        del.title = `DELETE /messages/${m.id}`;
        del.addEventListener("click", async () => {
          await fetch(`/messages/${m.id}`, { method: "DELETE" });
          refreshMessages();
        });
        li.append(when, text, del);
        return li;
      })
    );
  } catch {
    status.textContent = "Сервер не отвечает.";
  }
}

$("#message-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const input = $("#message-input");
  if (!input.value.trim()) return;
  await fetch("/messages", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message: input.value }),
  });
  input.value = "";
  refreshMessages();
});

/* ---------- запуск ---------- */

$("#recheck").addEventListener("click", runStageChecks);
refreshHealth();
runStageChecks();
setInterval(refreshHealth, 3000);
setInterval(refreshMessages, 3000);
