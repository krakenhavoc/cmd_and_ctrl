// Load test: ramp bot-only Commander tables against a cmd_and_ctrl server.
// Findings and how to run it safely: docs/environments.md, "Capacity".
//
// usage: node scripts/loadtest.mjs <base-url> <token-file | env:VAR>
//        [--steps 5,10,20] [--step-secs 180] [--bots 4] [--spectators 3]
//        [--max-p95-ms 2000] [--max-err 0.05] [--dry]
// The token is the admin token (CMDCTRL_ADMIN_TOKEN), read from a file or
// an environment variable and never printed. Every table is named
// loadtest-<run>-<n> and deleted at the end. Stop early: create the file
// named by LOADTEST_STOP (default ./loadtest.STOP), or press Ctrl-C.
// The run's log is written as loadtest-<run>.jsonl in the current directory.
import { readFileSync, existsSync, appendFileSync } from "node:fs";

const [base, tokenFile, ...rest] = process.argv.slice(2);
const opt = (k, d) => {
  const i = rest.indexOf(`--${k}`);
  return i >= 0 ? rest[i + 1] : d;
};
const STEPS = opt("steps", "5,10,20,40").split(",").map(Number);
const STEP_SECS = +opt("step-secs", "180");
const BOTS = +opt("bots", "4");
const SPECS = +opt("spectators", "3");
const MAX_P95 = +opt("max-p95-ms", "2000");
const MAX_ERR = +opt("max-err", "0.05");
const RUN = `loadtest-${new Date().toISOString().slice(5, 16).replace(/[-:T]/g, "")}`;
const LOG = `${RUN}.jsonl`;
const STOP = process.env.LOADTEST_STOP || "loadtest.STOP";
const log = (o) => {
  const line = JSON.stringify({ t: new Date().toISOString(), ...o });
  appendFileSync(LOG, line + "\n");
  console.log(line);
};

const adminSecret = tokenFile.startsWith("env:")
  ? (process.env[tokenFile.slice(4)] ?? "").trim()
  : readFileSync(tokenFile, "utf8").trim();
if (!adminSecret) {
  console.error(`no admin token in ${tokenFile}`);
  process.exit(1);
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
let token;
const stats = { req: 0, err: 0, codes: {}, lat: [] };
async function api(method, path, body, opts = {}) {
  for (let attempt = 0; ; attempt++) {
    const r = await api1(method, path, body, opts);
    if (r.status !== 429 || attempt >= 8) return r;
    stats.throttled = (stats.throttled || 0) + 1;
    await sleep(1000 * (attempt + 1));
  }
}
async function api1(method, path, body, { track = true } = {}) {
  const t0 = performance.now();
  let res, text;
  try {
    res = await fetch(base + path, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
      signal: AbortSignal.timeout(15000),
    });
    text = await res.text();
  } catch (e) {
    if (track) {
      stats.req++;
      stats.err++;
      stats.codes.net = (stats.codes.net || 0) + 1;
    }
    return { ok: false, status: 0, text: String(e) };
  }
  const ms = performance.now() - t0;
  if (track) {
    stats.req++;
    stats.lat.push(ms);
    if (res.status !== 429) stats.codes[res.status] = (stats.codes[res.status] || 0) + 1;
    if (!res.ok && res.status !== 429) stats.err++;
  }
  let json;
  try {
    json = JSON.parse(text);
  } catch {}
  return { ok: res.ok, status: res.status, json, text, ms };
}

const tables = new Map(); // id -> { sockets, frames, lastFrame, created }
const wsStats = { opened: 0, failed: 0, closed: 0, frames: 0 };
let decks = [];

function openSpectator(id) {
  const t = tables.get(id);
  if (!t) return;
  const ws = new WebSocket(
    `${base.replace(/^http/, "ws")}/ws?game=${id}&token=${encodeURIComponent(token)}`,
  );
  ws.onopen = () => {
    wsStats.opened++;
  };
  ws.onmessage = () => {
    wsStats.frames++;
    t.frames++;
    t.lastFrame = Date.now();
  };
  ws.onerror = () => {
    wsStats.failed++;
  };
  ws.onclose = () => {
    wsStats.closed++;
  };
  t.sockets.push(ws);
}

async function addTable(i) {
  const g = await api("POST", "/games", { name: `${RUN}-${i}` });
  if (!g.ok) return log({ ev: "create_fail", status: g.status, body: g.text.slice(0, 160) });
  const id = g.json.id;
  tables.set(id, { sockets: [], frames: 0, lastFrame: Date.now(), created: Date.now() });
  for (let b = 0; b < BOTS; b++) {
    const r = await api("POST", `/games/${id}/seats/bot`, {
      tier: "heuristic",
      deck: decks[(i + b) % decks.length],
      name: `LT ${b + 1}`,
    });
    if (!r.ok) log({ ev: "bot_fail", id, status: r.status, body: r.text.slice(0, 160) });
  }
  const s = await api("POST", `/games/${id}/start`);
  if (!s.ok) log({ ev: "start_fail", id, status: s.status, body: s.text.slice(0, 160) });
  for (let k = 0; k < SPECS; k++) openSpectator(id);
}

async function removeTable(id) {
  const t = tables.get(id);
  if (!t) return;
  for (const ws of t.sockets)
    try {
      ws.close();
    } catch {}
  tables.delete(id);
  let d;
  for (let i = 0; i < 60; i++) {
    d = await api("DELETE", `/games/${id}`, undefined, { track: false });
    if (d.ok || d.status === 404 || (d.status >= 400 && d.status < 500)) break;
    await sleep(10000); // 0 or 5xx: the server is down or restarting
  }
  if (!d.ok) {
    const a = await api("POST", `/games/${id}/archive`, undefined, { track: false });
    log({ ev: "cleanup", id, deleted: false, archived: a.ok, status: d.status });
  }
}

let cleaning = false;
async function cleanup(reason) {
  if (cleaning) return;
  cleaning = true;
  log({ ev: "cleanup_start", reason, tables: tables.size });
  for (const t of tables.values())
    for (const ws of t.sockets)
      try {
        ws.close();
      } catch {}
  // After a crash the server restores these tables on boot; wait for it, then delete them.
  for (let i = 0; i < 120; i++) {
    const h = await api1("GET", "/healthz", undefined, { track: false });
    if (h.ok) break;
    if (i % 6 === 0) log({ ev: "waiting_for_server", status: h.status });
    await sleep(5000);
  }
  const relog = await api1("POST", "/admin/login", { token: adminSecret }, { track: false });
  if (relog.ok) token = relog.json.token;
  for (const id of [...tables.keys()]) await removeTable(id);
  // Catch any table this run created that the map lost.
  const live = await api("GET", "/admin/live", undefined, { track: false });
  for (const t of live.json?.tables ?? [])
    if (t.name?.startsWith(RUN)) {
      await api("DELETE", `/games/${t.id}`, undefined, { track: false });
    }
  log({ ev: "cleanup_done" });
  process.exit(0);
}
process.on("SIGINT", () => cleanup("sigint"));
process.on("SIGTERM", () => cleanup("sigterm"));

const pct = (a, p) => {
  if (!a.length) return null;
  const s = [...a].sort((x, y) => x - y);
  return Math.round(s[Math.min(s.length - 1, Math.floor(p * s.length))]);
};

async function probe() {
  const h = [];
  for (let i = 0; i < 5; i++) {
    const r = await api("GET", "/healthz", undefined, { track: false });
    h.push(r.ok ? r.ms : 15000);
  }
  const live = await api("GET", "/admin/live", undefined, { track: false });
  const ended = (live.json?.tables ?? [])
    .filter((t) => t.name?.startsWith(RUN) && t.state !== "active" && t.state !== "lobby")
    .map((t) => t.id);
  const now = Date.now();
  const stalled = [...tables.values()].filter(
    (t) => now - t.created > 30000 && now - t.lastFrame > 30000,
  ).length;
  return {
    healthP50: pct(h, 0.5),
    healthP95: pct(h, 0.95),
    liveMs: Math.round(live.ms ?? -1),
    liveTotals: live.json?.totals,
    ended,
    stalled,
  };
}

async function main() {
  const login = await api("POST", "/admin/login", { token: adminSecret }, { track: false });
  if (!login.ok) {
    console.error("admin login failed", login.status);
    process.exit(1);
  }
  token = login.json.token;
  const d = await api("GET", "/decks", undefined, { track: false });
  decks = (d.json?.decks ?? []).map((x) => x.id);
  log({
    ev: "start",
    run: RUN,
    base,
    steps: STEPS,
    stepSecs: STEP_SECS,
    bots: BOTS,
    spectators: SPECS,
    decks: decks.length,
  });
  if (rest.includes("--dry")) return cleanup("dry");
  const base0 = await probe();
  log({ ev: "baseline", ...base0 });
  let made = 0;
  for (const target of STEPS) {
    log({ ev: "step", target });
    let createFails = 0;
    while (tables.size < target) {
      const before = tables.size;
      await addTable(made++);
      createFails = tables.size > before ? 0 : createFails + 1;
      if (createFails >= 3) {
        log({
          ev: "ceiling",
          target,
          tables: tables.size,
          reason: "3 table creates failed in a row",
        });
        return cleanup("ceiling");
      }
      await sleep(2600);
      if (existsSync(STOP)) return cleanup("stop-file");
    }
    const stepEnd = Date.now() + STEP_SECS * 1000;
    let lastFrames = wsStats.frames,
      lastT = Date.now();
    while (Date.now() < stepEnd) {
      await sleep(10000);
      if (existsSync(STOP)) return cleanup("stop-file");
      const p = await probe();
      for (const id of p.ended) {
        await removeTable(id);
        await addTable(made++);
      }
      const fps = (wsStats.frames - lastFrames) / ((Date.now() - lastT) / 1000);
      lastFrames = wsStats.frames;
      lastT = Date.now();
      const errRate = stats.req ? stats.err / stats.req : 0;
      const row = {
        ev: "sample",
        target,
        tables: tables.size,
        ...p,
        apiP95: pct(stats.lat.slice(-200), 0.95),
        errRate: +errRate.toFixed(3),
        codes: stats.codes,
        throttled: stats.throttled || 0,
        ws: { ...wsStats },
        framesPerSecPerTable: +(fps / Math.max(1, tables.size)).toFixed(2),
        replaced: p.ended.length,
      };
      delete row.ended;
      log(row);
      const why =
        p.healthP95 > MAX_P95
          ? `healthz p95 ${p.healthP95}ms > ${MAX_P95}`
          : errRate > MAX_ERR && stats.req > 20
            ? `error rate ${errRate.toFixed(3)}`
            : p.stalled > Math.max(2, tables.size * 0.2)
              ? `${p.stalled} tables stalled 30s+`
              : null;
      if (why) {
        log({ ev: "ceiling", target, reason: why });
        return cleanup("ceiling");
      }
    }
  }
  return cleanup("steps-complete");
}
main().catch((e) => {
  log({ ev: "crash", err: String(e) });
  cleanup("crash");
});
