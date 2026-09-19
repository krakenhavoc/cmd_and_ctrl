// @vitest-environment jsdom
//
// #740 — the board froze once with `state_unsafe_mutation` in a
// Board.svelte-owned effect, during the player's own attack
// declaration, with the REPLAY dock open after a failed 403 load. No
// stack was captured, and three manual attempts did not reproduce it.
//
// This file drives that sequence for real: the whole route, a live
// socket, the dock open on a 403, a hover over the creature being
// declared, and the attack-declaration frame — with a window-level
// error hook installed BEFORE any of it, which is the thing the
// original session was missing.
//
// The freeze itself is #266's mechanism. A throw that escapes a
// svelte/store subscriber leaves svelte/store's module-global
// `subscriber_queue` non-empty, and every later `set()` on every store
// in the app enqueues into a queue nobody drains — including
// `snapshot.set()`, which is why the board sat on one seq while the
// socket kept delivering frames. `guardedStore.ts` (#720) contains
// that; what follows pins the other half, that no such throw happens
// in the first place, and that the table keeps advancing.
//
// The root cause it is a regression test for is in
// hoverMeta.render.test.ts, which drives it in isolation.

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import { recentClientErrors, resetClientErrors, installErrorCapture } from "./clientErrors";
import { loadAppConfig, resetAppConfigForTests } from "./env";
import { PROTOCOL_VERSION, type CardView, type GameView, type PlayerView } from "./protocol";
import { hoveredCard } from "./cardTypes";
import { session } from "./session";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

const Game = await import("../routes/Game.svelte").then((m) => m.default);

// ---------------------------------------------------------------- //
// The table: two seats, and two copies of one printing on the        //
// battlefield. Two copies matter — they share a `scryfall_id`, which //
// is the state the hover panel's metadata cache keys on.             //
// ---------------------------------------------------------------- //

const ME = "seat-me";
const OPP = "seat-opp";
const PRINTING = "sf-grizzly";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const bear = (id: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: "Grizzly Bears",
    owner: ME,
    controller: ME,
    known_by_you: true,
    scryfall_id: PRINTING,
    type_line: "Creature — Bear",
    power: 2,
    toughness: 2,
    ...extra,
  }) as CardView;

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

const gameView = (battlefield: CardView[]): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Katara", 0), seat(OPP, "Aang", 1)],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      number: 4,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
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

// ---------------------------------------------------------------- //
// Fakes                                                              //
// ---------------------------------------------------------------- //

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

// routeFetch is the whole network for this test. The replay endpoint
// answers 403 — "players can't load a replay until the game ends",
// which is the state the dock was left in.
let replayCalls = 0;
function routeFetch(input: unknown): Promise<Response> {
  const url = String(typeof input === "string" ? input : ((input as { url?: string })?.url ?? ""));
  if (url.includes("/config") && !url.includes("bugreport")) {
    return Promise.resolve({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          env: "dev",
          features: { replay_scrubber: true, frame_inspector: false },
        }),
    } as unknown as Response);
  }
  if (url.includes("/replay")) {
    replayCalls += 1;
    const body = { error: "replays are available once the game is over" };
    const fail = {
      ok: false,
      status: 403,
      statusText: "Forbidden",
      json: () => Promise.resolve(body),
      // authFetch reads the body off a clone so the caller can still
      // have it; a stub without one reports "403 undefined".
      clone: () => ({ json: () => Promise.resolve(body) }),
    };
    return Promise.resolve(fail as unknown as Response);
  }
  if (url.includes("/cards/")) {
    return Promise.resolve({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          id: PRINTING,
          name: "Grizzly Bears",
          type_line: "Creature — Bear",
          oracle_text: "",
        }),
    } as unknown as Response);
  }
  // Everything else (the bug-report probe) fails the way the network
  // does when the feature is off.
  return Promise.reject(new Error("no network in this test"));
}

let realWebSocket: unknown;
let realFetch: unknown;
const uncaught: string[] = [];

function onWindowError(ev: Event): void {
  uncaught.push(String((ev as ErrorEvent).message ?? ev.type));
}

// jsdom ships neither of these and the board measures itself with
// both. Stubs, not mocks: nothing here is under test, and a board that
// cannot measure itself is not the board that froze.
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
  // jsdom's `play()` returns undefined and logs "not implemented";
  // sounds.ts calls `.catch()` on the promise the DOM actually
  // returns, so without this the board throws on the tap sound and
  // the boundary swallows the very sequence under test.
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
  replayCalls = 0;
  uncaught.length = 0;
  resetClientErrors();
  resetAppConfigForTests();
  hoveredCard.set(null);
  session.set(null);
  // The hook the original session did not have, installed before a
  // line of app code runs.
  installErrorCapture(window, console);
  window.addEventListener("error", onWindowError);
});

afterEach(() => {
  cleanup();
  window.removeEventListener("error", onWindowError);
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
});

// settle lets the queued fetches (config, card metadata) resolve and
// the resulting effects flush.
async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    await Promise.resolve();
    flushSync();
  }
}

async function mountGame(): Promise<{ container: HTMLElement; socket: FakeSocket }> {
  // The route reads the dev-feature flags off the shared config store,
  // and the dock only mounts when one of them is on.
  await loadAppConfig();
  const handle = render(Game as never, { gameID: "g1" } as never);
  const socket = FakeSocket.last;
  expect(socket, "Game.svelte should have opened a socket on mount").toBeTruthy();
  socket!.emit("open", {});
  socket!.emit("message", { data: snapshotFrame(165, gameView([bear("bear-1"), bear("bear-2")])) });
  await settle();
  return { container: handle.container, socket: socket! };
}

const buttonSaying = (c: ParentNode, re: RegExp): HTMLElement | undefined =>
  [...c.querySelectorAll<HTMLElement>("button")].find((b) => re.test(b.textContent ?? ""));

const unsafeMutations = (): string[] =>
  recentClientErrors()
    .map((e) => e.text)
    .filter((t) => t.includes("state_unsafe_mutation"));

describe("#740 — the attack declaration with the replay dock open", () => {
  it("opens the replay dock and leaves it on a failed 403 load", async () => {
    const { container } = await mountGame();

    const replayTab = buttonSaying(container, /^\s*REPLAY\s*$/i);
    expect(replayTab, "the dev dock should offer a REPLAY tab").toBeTruthy();
    click(replayTab!);
    await settle();

    const load = buttonSaying(container, /load replay/i);
    expect(load, "the replay panel should offer a load button").toBeTruthy();
    click(load!);
    await settle();

    expect(replayCalls, "the replay endpoint should have been called").toBe(1);
    const alert = container.querySelector('[role="alert"]');
    expect(alert?.textContent ?? "", "the 403 should be shown in the dock").toMatch(/replay/i);
  });

  it("keeps drawing and keeps applying frames through the whole sequence", async () => {
    const { container, socket } = await mountGame();

    // 1. The dock, open on a failed load, exactly as the session that
    //    froze had left it.
    click(buttonSaying(container, /^\s*REPLAY\s*$/i)!);
    await settle();
    click(buttonSaying(container, /load replay/i)!);
    await settle();
    expect(replayCalls).toBe(1);

    // 2. The player looks at the creature she is about to attack with,
    //    and then at her other copy of it. Two permanents, one
    //    printing, one metadata store — the shape that made a
    //    `$derived` write to a subscribed store.
    const tiles = [...container.querySelectorAll<HTMLElement>(".card")];
    expect(tiles.length, "the battlefield should have rendered its cards").toBeGreaterThan(0);
    hoveredCard.set(bear("bear-1"));
    await settle();
    hoveredCard.set(bear("bear-2"));
    await settle();

    // 3. Her own attack declaration arrives.
    const declared = gameView([
      bear("bear-1", { attacking_target: OPP, attacking_target_kind: "player", tapped: true }),
      bear("bear-2"),
    ]);
    socket.emit("message", { data: snapshotFrame(166, declared) });
    await settle();

    // The board MOVED. This is the assertion the bug report is about:
    // it stayed at seq 165 while frames kept arriving.
    expect(
      container.querySelector(".card.attacking"),
      "the declared attacker should be drawn as attacking",
    ).toBeTruthy();
    expect(
      container.querySelector(".board-failed"),
      "the boundary fallback should not be showing",
    ).toBeNull();
    expect(unsafeMutations(), "no state_unsafe_mutation anywhere in the sequence").toEqual([]);
    expect(uncaught, "no uncaught window error").toEqual([]);

    // And a further frame still lands, which is what a poisoned
    // subscriber queue would have stopped.
    const cleared = gameView([bear("bear-1", { tapped: true }), bear("bear-2")]);
    socket.emit("message", { data: snapshotFrame(167, cleared) });
    await settle();
    expect(unsafeMutations()).toEqual([]);
    expect(uncaught).toEqual([]);
  });
});
