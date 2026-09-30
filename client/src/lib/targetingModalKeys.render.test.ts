// @vitest-environment jsdom
//
// #1659 — Game.svelte's window-level Escape/Enter handler (cancel /
// confirm the active targeting walk) used to fire even when a modal
// opened during that same cast — the mode picker, an X prompt, a
// sacrifice/discard cost picker, DivideDamageModal, an alt-cost
// picker, ChoicePromptModal. Escape closed the modal AND cancelled
// the whole cast; Enter confirmed the modal AND completed the
// targeting walk underneath it.
//
// The fix reads the same `modalOpen` store every modal already
// registers a layer on (lib/modalLayers.ts) via `<ModalLayer />`, so
// this test drives that store directly with `pushModalLayer()` rather
// than mounting one specific modal — the behaviour under test is
// Game.svelte's handler standing down while ANY layer is registered,
// which is the whole point of a single shared source of truth.

import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { get } from "svelte/store";

import { loadAppConfig, resetAppConfigForTests } from "./env";
import { PROTOCOL_VERSION, type CardView, type GameView, type PlayerView } from "./protocol";
import { session } from "./session";
import { targeting, begin, setConfirmHandler } from "./targeting";
import { pushModalLayer, _resetForTests as resetModalLayers } from "./modalLayers";
import { render, cleanup, flushSync } from "./test/render.svelte";

const Game = await import("../routes/Game.svelte").then((m) => m.default);

const ME = "seat-me";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, name: string, n: number): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Katara", 0)],
    battlefield: zone("battlefield", undefined, []),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "main",
    },
    mulligans_open: false,
  }) as unknown as GameView;

const snapshotFrame = (seq: number, view: GameView): string =>
  JSON.stringify({
    v: PROTOCOL_VERSION,
    kind: "snapshot",
    id: `frame-${seq}`,
    payload: { seq, game: view },
  });

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;

  readyState = FakeSocket.OPEN;
  sent: string[] = [];
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};

  constructor(readonly url: string) {
    FakeSocket.last = this;
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(data: string): void {
    this.sent.push(data);
  }
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

function routeFetch(input: unknown): Promise<Response> {
  const url = String(typeof input === "string" ? input : ((input as { url?: string })?.url ?? ""));
  if (url.includes("/config") && !url.includes("bugreport")) {
    return Promise.resolve({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ env: "dev", features: {} }),
    } as unknown as Response);
  }
  return Promise.reject(new Error("no network in this test"));
}

let realWebSocket: unknown;
let realFetch: unknown;

class FakeResizeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  g.fetch = routeFetch;
  g.ResizeObserver ??= FakeResizeObserver;
  g.IntersectionObserver ??= FakeResizeObserver;
  HTMLMediaElement.prototype.play = () => Promise.resolve();
  if (typeof window.matchMedia !== "function") {
    window.matchMedia = ((q: string) => ({
      matches: false,
      media: q,
      addEventListener() {},
      removeEventListener() {},
      addListener() {},
      removeListener() {},
      onchange: null,
      dispatchEvent: () => false,
    })) as unknown as typeof window.matchMedia;
  }
  FakeSocket.last = null;
  resetAppConfigForTests();
  session.set(null);
  targeting.set(null);
  setConfirmHandler(null);
  resetModalLayers();
});

afterEach(() => {
  cleanup();
  targeting.set(null);
  setConfirmHandler(null);
  resetModalLayers();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
});

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    await Promise.resolve();
    flushSync();
  }
}

async function mountGame(): Promise<void> {
  await loadAppConfig();
  render(Game as never, { gameID: "g1" } as never);
  const socket = FakeSocket.last;
  expect(socket, "Game.svelte should have opened a socket on mount").toBeTruthy();
  socket!.emit("open", {});
  socket!.emit("message", { data: snapshotFrame(1, gameView()) });
  await settle();
}

const card = (name: string): CardView => ({
  instance_id: name.toLowerCase(),
  name,
  owner: ME,
  controller: ME,
});

function pressKey(key: string): void {
  window.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true }));
}

describe("#1659 — the global Escape/Enter handler stands down behind a modal", () => {
  it("without a modal open, Escape cancels the active targeting walk", async () => {
    await mountGame();
    begin(card("Lightning Bolt"), "any");
    expect(get(targeting)).not.toBeNull();

    pressKey("Escape");

    expect(get(targeting)).toBeNull();
  });

  it("without a modal open, Enter confirms the active targeting walk", async () => {
    await mountGame();
    begin(card("Lightning Bolt"), "any");
    let confirmed = 0;
    setConfirmHandler(() => {
      confirmed += 1;
    });

    pressKey("Enter");

    expect(confirmed).toBe(1);
  });

  it("with a modal open, Escape does NOT also cancel the targeting walk underneath it", async () => {
    await mountGame();
    const unregister = pushModalLayer();
    begin(card("Lightning Bolt"), "any");
    expect(get(targeting)).not.toBeNull();

    pressKey("Escape");

    // The modal's own Escape handler owns this keypress; Game.svelte's
    // global handler must not have also cancelled the walk.
    expect(get(targeting)).not.toBeNull();

    unregister();
  });

  it("with a modal open, Enter does NOT also confirm the targeting walk underneath it", async () => {
    await mountGame();
    const unregister = pushModalLayer();
    begin(card("Lightning Bolt"), "any");
    let confirmed = 0;
    setConfirmHandler(() => {
      confirmed += 1;
    });

    pressKey("Enter");

    expect(confirmed).toBe(0);

    // Once the modal closes, the same keypress works normally again —
    // this isn't a stuck-off switch, it tracks the layer.
    unregister();
    pressKey("Enter");
    expect(confirmed).toBe(1);
  });
});
