// @vitest-environment jsdom
//
// groupAttackStage.render.test.ts — #4423, through the real Game route.
// Choosing a defender for a token group ("Attack Bot 1 with 3" in the
// group's list) used to send the group as its own declaration, which
// emptied the staged set: the creatures the player had clicked one by
// one were dropped, and the group's count could not be changed. A
// group's attack is now staged like any other click, joins the set, and
// goes out with it as one declare_attackers. The board is a stub that
// presses the group's buttons.

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

describe("a token group's attack joins the staged set (#4423)", () => {
  it("keeps the creature clicked one by one, and sends both in one declaration", async () => {
    const c = await mountGame(declareAttackers());
    pressed(c, "select-b1");
    expect(staged(c)).toBe("b1");

    pressed(c, "group-attack");
    expect(staged(c)).toBe("b1,w1,w2,w3");
    expect(FakeSocket.last!.declarations()).toHaveLength(0);
    expect(dockButtons(c, /^Attack with 4$/)).toHaveLength(1);

    click(dockButtons(c, /^Attack with 4$/)[0]!);
    const sent = FakeSocket.last!.declarations();
    expect(sent).toHaveLength(1);
    expect(sent[0]!.payload!.params!.attackers!.map((a) => a.attacker)).toEqual([
      "b1",
      "w1",
      "w2",
      "w3",
    ]);
  });

  it("lets the group's count change after it was chosen", async () => {
    const c = await mountGame(declareAttackers());
    pressed(c, "select-b1");
    pressed(c, "group-attack");
    pressed(c, "group-unstage");
    expect(staged(c)).toBe("b1,w1,w2");
    expect(dockButtons(c, /^Attack with 3$/)).toHaveLength(1);
    // And more can be added again.
    pressed(c, "group-attack");
    expect(staged(c)).toBe("b1,w1,w2,w3");
  });
  it("keeps a click-staged creature when the group's exert question opens the picker", async () => {
    // b1 can only attack Cy, so one click stages it there; the wolves
    // attack Opp, and w1 may be exerted, which asks in the picker.
    const c = await mountGame(
      declareAttackers({ b1: { attack_targets: [OPP2] }, w1: { exert_on_attack: true } }, true),
    );
    pressed(c, "select-b1");
    expect(staged(c)).toBe("b1");

    pressed(c, "group-attack");
    expect(FakeSocket.last!.declarations()).toHaveLength(0);
    // The picker lists the wolves (b1 cannot attack Opp) and confirms them.
    const confirm = dockButtons(c, /^Attack with 3/);
    expect(confirm).toHaveLength(1);
    click(confirm[0]!);

    const sent = FakeSocket.last!.declarations();
    expect(sent).toHaveLength(1);
    expect(sent[0]!.payload!.params!.attackers!.map((a) => [a.attacker, a.target])).toEqual([
      ["w1", OPP],
      ["w2", OPP],
      ["w3", OPP],
      ["b1", OPP2],
    ]);
  });
});
