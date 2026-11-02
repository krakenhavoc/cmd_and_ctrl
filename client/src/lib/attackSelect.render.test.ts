// @vitest-environment jsdom
//
// attackSelect.render.test.ts — #4326, through the real Game route. At a
// table where each creature can attack more than one seat, a click
// selects a creature and a click on a seat stages it there. A click on a
// second creature now joins the selection, and one seat click stages all
// of them. The board is the token-group stub, which presses the same
// callbacks a click on a creature or a player does.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import { PROTOCOL_VERSION, type CardView, type GameView, type PlayerView } from "./protocol";
import { session, type Session } from "./session";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/BoardGroupAttackStub.svelte")).default,
}));

const Game = await import("../routes/Game.svelte").then((m) => m.default);

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

interface Frame {
  kind: string;
  payload?: { type?: string; params?: { attackers?: Array<{ attacker: string; target: string }> } };
}

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  sent: Frame[] = [];
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.last = this;
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(data: string): void {
    this.sent.push(JSON.parse(data));
  }
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
  declarations() {
    return this.sent.filter((f) => f.kind === "action" && f.payload?.type === "declare_attackers");
  }
}

const ME = "p-1";
const OPP = "p-2";
const OPP2 = "p-3";

const zone = (kind: string, owner?: string, cards: CardView[] = []) => ({
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
    hand_kept: true,
    undos_remaining: 1,
  }) as unknown as PlayerView;
const creature = (id: string, name: string): CardView =>
  ({
    instance_id: id,
    name,
    controller: ME,
    owner: ME,
    type_line: "Creature",
  }) as CardView;

const ATTACKERS = ["b1", "w1", "w2", "w3"];

function declareAttackers(
  over: Record<string, Record<string, unknown>> = {},
  extraSeat = false,
): GameView {
  const sources = Object.fromEntries(
    ATTACKERS.map((id) => [
      id,
      { kinds: ["attack"], moves: 1, attack_targets: [OPP], ...over[id] },
    ]),
  );
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Opp", 1), ...(extraSeat ? [seat(OPP2, "Cy", 2)] : [])],
    battlefield: zone("battlefield", undefined, [
      creature("b1", "Grizzly Bears"),
      creature("w1", "Wolf"),
      creature("w2", "Wolf"),
      creature("w3", "Wolf"),
    ]),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    legal_actions: { pass: true, sources },
    turn: {
      seq: 3,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
    },
    mulligans_open: false,
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
  resetDock();
  resetModals();
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  session.set(playerSession());
});

afterEach(() => {
  cleanup();
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

const pressed = (c: ParentNode, id: string): void =>
  click(c.querySelector<HTMLButtonElement>(`[data-testid="${id}"]`)!);
const staged = (c: ParentNode): string =>
  c.querySelector('[data-testid="staged"]')!.textContent ?? "";
const dockButtons = (c: ParentNode, name: RegExp): HTMLButtonElement[] =>
  [
    ...c.querySelectorAll<HTMLButtonElement>('.play-area > section[aria-label="actions"] button'),
  ].filter((b) => name.test((b.getAttribute("aria-label") ?? b.textContent ?? "").trim()));

const selected = (c: ParentNode): string =>
  c.querySelector('[data-testid="selected"]')!.textContent ?? "";
const dockQuestion = (c: ParentNode): string =>
  c.querySelector('.play-area > section[aria-label="actions"]')!.textContent ?? "";

// Every creature can attack Opp or Cy.
const fourWay = () =>
  declareAttackers(
    Object.fromEntries(ATTACKERS.map((id) => [id, { attack_targets: [OPP, OPP2] }])),
    true,
  );

describe("selecting several attackers, then one target (#4326)", () => {
  it("stages every selected creature at the one seat clicked, in one declaration", async () => {
    const c = await mountGame(fourWay());
    pressed(c, "select-w1");
    pressed(c, "select-w2");
    pressed(c, "select-w3");
    expect(staged(c)).toBe("");
    expect(selected(c)).toBe("w3");
    expect(dockQuestion(c)).toContain("Attacking with Wolf and 2 more");

    pressed(c, "target-p-2");
    expect(staged(c)).toBe("w1,w2,w3");
    expect(selected(c)).toBe("");
    expect(FakeSocket.last!.declarations()).toHaveLength(0);

    click(dockButtons(c, /^Attack with 3/)[0]!);
    const sent = FakeSocket.last!.declarations();
    expect(sent).toHaveLength(1);
    expect(sent[0]!.payload!.params!.attackers!.map((a) => [a.attacker, a.target])).toEqual([
      ["w1", OPP],
      ["w2", OPP],
      ["w3", OPP],
    ]);
  });

  it("takes one creature out of the selection on its second click", async () => {
    const c = await mountGame(fourWay());
    pressed(c, "select-w1");
    pressed(c, "select-w2");
    pressed(c, "select-w3");
    // The last one clicked: the one before it takes its place.
    pressed(c, "select-w3");
    expect(selected(c)).toBe("w2");
    // An earlier one.
    pressed(c, "select-w1");
    expect(selected(c)).toBe("w2");
    pressed(c, "target-p-3");
    expect(staged(c)).toBe("w2");
  });

  it("splits an attack: a second batch goes at another seat", async () => {
    const c = await mountGame(fourWay());
    pressed(c, "select-w1");
    pressed(c, "select-w2");
    pressed(c, "target-p-2");
    pressed(c, "select-w3");
    pressed(c, "select-b1");
    pressed(c, "target-p-3");
    click(dockButtons(c, /^Attack with 4/)[0]!);
    const sent = FakeSocket.last!.declarations();
    expect(sent[0]!.payload!.params!.attackers!.map((a) => [a.attacker, a.target])).toEqual([
      ["w1", OPP],
      ["w2", OPP],
      ["w3", OPP2],
      ["b1", OPP2],
    ]);
  });

  it("leaves a creature that can't attack the seat selected", async () => {
    const c = await mountGame(
      declareAttackers(
        {
          b1: { attack_targets: [OPP2, "pw-1"] },
          w1: { attack_targets: [OPP, OPP2] },
        },
        true,
      ),
    );
    pressed(c, "select-b1");
    pressed(c, "select-w1");
    pressed(c, "target-p-2");
    expect(staged(c)).toBe("w1");
    expect(selected(c)).toBe("b1");
  });

  it("selects a creature with an exert question alone", async () => {
    const c = await mountGame(
      declareAttackers(
        {
          w1: { attack_targets: [OPP, OPP2] },
          w2: { attack_targets: [OPP, OPP2], exert_on_attack: true },
        },
        true,
      ),
    );
    pressed(c, "select-w1");
    pressed(c, "select-w2");
    expect(selected(c)).toBe("w2");
    expect(dockQuestion(c)).not.toContain("and 1 more");
    // Cancel clears it.
    click(dockButtons(c, /^Cancel/)[0]!);
    expect(selected(c)).toBe("");
  });
});
