// @vitest-environment jsdom
//
// tableHints.render.test.ts — the table's first-use hints on the real
// Game route (ADR 0125 §3.7, Delivery PR 7): each hint's anchor resolves
// on a mounted table where its feature is on screen, the opening roll's
// and the stack's anchors are absent when there is no roll or stack, and
// the table publishes a moment that is not quiet while the viewer owes
// the opening roll, which is what keeps every table hint away from it.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

import { PROTOCOL_VERSION, type CardView, type GameView, type PlayerView } from "../protocol";
import { session, type Session } from "../session";
import { _resetForTests as resetDock } from "../dock";
import { _resetForTests as resetModals } from "../modalLayers";
import { targeting, setConfirmHandler } from "../targeting";
import { defaultSettings, settings } from "../settings";
import { cleanup, flushSync, render } from "../test/render.svelte";
import { resolveAnchor } from "../tutorialAnchor";
import { L } from "../labels";
import { HINTS } from "./index";
import { anchorOf, emptyContext, type Hint, type HintContext } from "./hint";
import { _resetTableMomentForTests, notQuietReason, tableState } from "./tableMoment";

// Every test here mounts the full Game route in jsdom, which takes 1-4 s on a
// loaded runner (#2508). The 5 s default left no headroom, so this file, and
// only this file, gets a longer limit; the global default is untouched.
vi.setConfig({ testTimeout: 20_000 });

const Game = await import("../../routes/Game.svelte").then((m) => m.default);

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.last = this;
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(): void {}
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

const ME = "p-1";
const zone = (kind: string, owner?: string, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const commander = (): CardView =>
  ({
    instance_id: "cmd",
    name: "Atraxa",
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Legendary Creature — Phyrexian Angel",
  }) as unknown as CardView;

const relic = (): CardView =>
  ({
    instance_id: "relic",
    name: "Relic of Sauron",
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Legendary Artifact",
    activated_abilities: [{ index: 0, ref: "own:0", label: "{1}: Scry 1.", mana_cost: "{1}" }],
  }) as unknown as CardView;

const seat = (id: string, name: string, n: number, cmd: CardView[] = []): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id, cmd),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    hand_kept: true,
    undos_remaining: 1,
  }) as unknown as PlayerView;

interface Opts {
  stack?: boolean;
  roll?: boolean;
}

/** A four-seat table at the viewer's third turn, or in the opening roll. */
function table(o: Opts = {}): GameView {
  const spell = { ...relic(), instance_id: "spell", name: "Shock" } as CardView;
  return {
    id: "g",
    state: "active",
    seats: [
      seat(ME, "Me", 0, [commander()]),
      seat("p-2", "Bob", 1),
      seat("p-3", "Cat", 2),
      seat("p-4", "Dee", 3),
    ],
    battlefield: zone("battlefield", undefined, [relic()]),
    stack: zone("stack", undefined, o.stack ? [spell] : []),
    exile: zone("exile"),
    stack_items: o.stack
      ? [{ id: "i1", kind: "spell", controller: ME, owner: ME, source_card_id: "spell" }]
      : [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 5,
      number: o.roll ? 0 : 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
    log: [],
    ...(o.roll
      ? { opening_roll: { rounds: [{ seats: [0, 1, 2, 3], rolls: [{ seat: 0, result: 12 }] }] } }
      : {}),
  } as unknown as GameView;
}

function playerSession(): Session {
  const expiresAt = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token: "tok",
    expiresAt,
    principal: {
      role: "player",
      game_id: "game-1",
      player_id: ME,
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    playerID: ME,
    gameID: "game-1",
  } as Session;
}

let realWebSocket: unknown;
let realFetch: unknown;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  g.fetch = () => Promise.resolve(new Response("{}", { status: 404 }));
  FakeSocket.last = null;
  settings.set(defaultSettings());
  resetDock();
  resetModals();
  _resetTableMomentForTests();
  targeting.set(null);
  setConfirmHandler(null);
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  vi.spyOn(HTMLMediaElement.prototype, "load").mockImplementation(() => {});
  session.set(playerSession());
});

afterEach(() => {
  cleanup();
  resetModals();
  _resetTableMomentForTests();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
  session.set(null);
  vi.restoreAllMocks();
});

async function mountGame(game: GameView): Promise<HTMLElement> {
  const handle = render(Game as never, { gameID: "game-1" } as never);
  FakeSocket.last!.emit("open", {});
  FakeSocket.last!.emit("message", {
    data: JSON.stringify({
      v: PROTOCOL_VERSION,
      kind: "snapshot",
      id: "frame-1",
      payload: { seq: 1, game },
    }),
  });
  flushSync();
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
  return handle.container;
}

const TABLE = HINTS.filter((h) => h.place === "table");
const hint = (id: string): Hint => TABLE.find((h) => h.id === id)!;

/** The hint's anchor, resolved in the DOM over the view the table shows. */
function resolved(id: string, view: GameView): Element | null {
  const ctx: HintContext = {
    ...emptyContext("table", "game"),
    view,
    viewerID: ME,
  };
  const a = anchorOf(hint(id), ctx);
  return a ? resolveAnchor(a) : null;
}

describe("the table's hint anchors on a mounted table", () => {
  it("resolve for the dock, the viewer's board, commander, abilities, the strip, ⋯ and a rival's board", async () => {
    const view = table();
    await mountGame(view);
    for (const id of [
      "table.dock",
      "table.right-click",
      "table.commander",
      "table.attention",
      "table.more",
      "table.expand",
      "table.shortcuts",
    ]) {
      expect(resolved(id, view), id).not.toBeNull();
    }
  });

  it("puts the commander's zone inside the viewer's own board", async () => {
    const view = table();
    await mountGame(view);
    const zoneEl = resolved("table.commander", view)!;
    expect(zoneEl.getAttribute("aria-label")).toMatch(/ command zone, /);
    expect(zoneEl.closest('[aria-label="your board"]')).not.toBeNull();
  });

  // One mount per test: mounting the whole Game route is the expensive
  // step (about 1.3 s each on a quiet machine, #2508), and the earlier
  // single test mounted twice and ran past the 5 s default on CI.
  it("finds the stack's pile while an item waits", async () => {
    const view = table({ stack: true });
    await mountGame(view);
    expect(resolved("table.stack", view)).not.toBeNull();
  });

  it("does not find the stack's pile when nothing is on the stack", async () => {
    const empty = table();
    await mountGame(empty);
    expect(resolved("table.stack", empty)).toBeNull();
  });

  it("finds the opening roll's banner only during the roll", async () => {
    const view = table();
    await mountGame(view);
    expect(resolved("table.opening-roll", view)).toBeNull();
  });
});

describe("the opening roll's hint", () => {
  it("has its banner once the viewer has rolled, and the table is quiet then", async () => {
    const view = table({ roll: true });
    await mountGame(view);
    expect(resolved("table.opening-roll", view)).not.toBeNull();
    const t = get(tableState);
    expect(t).not.toBeNull();
    // The viewer rolled, others have not: only a status line, no request.
    expect(notQuietReason({ ...t!.moment, since: 0 }, 60_000, null)).toBeNull();
  });

  it("is not offered while the viewer's own roll is the dock's request", async () => {
    const view = table({ roll: true });
    (view.opening_roll as { rounds: { rolls: unknown[] }[] }).rounds[0].rolls = [];
    await mountGame(view);
    const t = get(tableState);
    expect(t).not.toBeNull();
    expect(notQuietReason({ ...t!.moment, since: 0 }, 60_000, null)).toBe("dock-request");
  });
});

describe("a press on the tip card (#2422, #2372)", () => {
  function pointer(type: string, target: Element): void {
    target.dispatchEvent(new Event(type, { bubbles: true }));
    flushSync();
  }

  it("is not a gesture, so the card stays up for its own button's click", async () => {
    await mountGame(table());
    const card = document.createElement("aside");
    card.setAttribute("aria-label", L.tip);
    const button = document.createElement("button");
    button.textContent = "Got it";
    card.append(button);
    document.body.append(card);
    try {
      pointer("pointerdown", button);
      expect(get(tableState)!.moment.gesture).toBe(false);
      pointer("pointerup", button);
      // A press anywhere else on the table still is one.
      pointer("pointerdown", document.body);
      expect(get(tableState)!.moment.gesture).toBe(true);
      pointer("pointerup", document.body);
      expect(get(tableState)!.moment.gesture).toBe(false);
    } finally {
      card.remove();
    }
  });
});
