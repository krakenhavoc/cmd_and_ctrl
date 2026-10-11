// @vitest-environment jsdom
//
// combatDock.render.test.ts — ADR 0111 Delivery PR 3 (S56, #1958),
// through the real Game route. Combat left the attention strip for the
// action dock: the attack row in declare attackers, No blocks / Done
// blocking for a defender who owes blocks, and the combat selection's
// hint with Cancel. Undo left the ⋯ menu and the attack row: there is
// one, in the dock's toggles row. The board is a stub that renders the
// strip Game.svelte hands it, so these read what is (and is not) in it.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import { PROTOCOL_VERSION, type CardView, type GameView, type PlayerView } from "./protocol";
import { session, type Session } from "./session";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const Game = await import("../routes/Game.svelte").then((m) => m.default);

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  sent: Array<{ kind: string; payload?: { type?: string } }> = [];
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
  actions(): string[] {
    return this.sent.filter((f) => f.kind === "action").map((f) => f.payload?.type ?? "");
  }
}

const ME = "p-1";
const OPP = "p-2";

const zone = (kind: string, owner?: string, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});
const seat = (id: string, name: string, n: number, over: Partial<PlayerView> = {}): PlayerView =>
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
    ...over,
  }) as unknown as PlayerView;
const creature = (id: string, controller: string, over: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: "Grizzly Bears",
    controller,
    owner: controller,
    type_line: "Creature — Bear",
    ...over,
  }) as CardView;

function declareAttackers(over: Partial<PlayerView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, "Me", 0, over), seat(OPP, "Opp", 1)],
    battlefield: zone("battlefield", undefined, [creature("b1", ME), creature("b2", ME)]),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
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

function declareBlockers(staged: boolean): GameView {
  return {
    ...declareAttackers(),
    battlefield: zone("battlefield", undefined, [
      creature("a1", OPP, { attacking_target: ME } as Partial<CardView>),
      creature("b1", ME, staged ? ({ blocking_target: "a1" } as Partial<CardView>) : {}),
    ]),
    turn: {
      seq: 4,
      number: 3,
      active_seat: 1,
      priority_holder: 0,
      phase: "combat",
      step: "declare_blockers",
      block_pending_seats: [0],
      block_decision_seats: [0],
    },
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
let seq = 1;

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
  // jsdom's media play() returns nothing; the attack sound awaits it.
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

function snapshot(game: GameView): void {
  FakeSocket.last!.emit("message", {
    data: JSON.stringify({
      v: PROTOCOL_VERSION,
      kind: "snapshot",
      id: `frame-${seq}`,
      payload: { seq: seq++, game },
    }),
  });
  flushSync();
}

async function mountGame(game: GameView): Promise<HTMLElement> {
  const handle = render(Game as never, { gameID: "game-1" } as never);
  FakeSocket.last!.emit("open", {});
  snapshot(game);
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
  return handle.container;
}

const dockOf = (c: ParentNode) =>
  c.querySelector<HTMLElement>('.play-area > section[aria-label="actions"]')!;
const stripOf = (c: ParentNode) => c.querySelector<HTMLElement>('[data-testid="strip"]')!;
function accessibleName(el: Element): string {
  const label = el.getAttribute("aria-label");
  if (label) return label;
  const copy = el.cloneNode(true) as Element;
  copy.querySelectorAll('[aria-hidden="true"]').forEach((n) => n.remove());
  return (copy.textContent ?? "").replace(/\s+/g, " ").trim();
}
const buttons = (root: ParentNode, name: string | RegExp): HTMLButtonElement[] =>
  [...root.querySelectorAll<HTMLButtonElement>("button")].filter((b) =>
    typeof name === "string" ? accessibleName(b) === name : name.test(accessibleName(b)),
  );

describe("combat in the action dock", () => {
  it("puts the attack row in the dock in declare attackers, and nothing combat in the strip", async () => {
    const c = await mountGame(declareAttackers());
    const dock = dockOf(c);
    const row = dock.querySelector('[role="dialog"][aria-label="declare attackers"]')!;
    expect(row).not.toBeNull();
    const group = row.querySelector('[role="group"][aria-label="declare attackers"]')!;
    expect(group.textContent).toContain("2 ready to attack");
    const all = buttons(group, "Attack Opp with all 2 creatures");
    expect(all).toHaveLength(1);
    // The step continues with next: it is still in the bar.
    expect(buttons(dock, "next")).toHaveLength(1);

    const strip = stripOf(c);
    expect(strip.querySelector(".att.attack-all")).toBeNull();
    expect(buttons(strip, /attack|undo|block|cancel/i)).toHaveLength(0);
    expect(buttons(c, /^Attack Opp with all/)).toHaveLength(1);

    click(all[0]!);
    expect(FakeSocket.last!.actions()).toContain("declare_attackers");
  });

  it("keeps the row after a declaration, and the one Undo in the toggles row takes it back", async () => {
    const view = declareAttackers();
    view.battlefield.cards = [
      creature("b1", ME, { attacking_target: OPP } as Partial<CardView>),
      creature("b2", ME, { attacking_target: OPP } as Partial<CardView>),
    ];
    const c = await mountGame(view);
    const dock = dockOf(c);
    const group = dock.querySelector('[role="group"][aria-label="declare attackers"]')!;
    expect(group.textContent).toContain("2 declared");
    // No Undo in the attack row any more…
    expect(buttons(group, /undo/i)).toHaveLength(0);
    // …one on the whole page, in the dock's toggles row.
    const undos = buttons(c, /undo/i);
    expect(undos).toHaveLength(1);
    expect(undos[0]!.closest('[role="group"]')?.getAttribute("aria-label")).toBe(
      "priority controls",
    );
    expect(undos[0]!.getAttribute("aria-label")).toBe("Undo (1 left)");
    click(undos[0]!);
    expect(FakeSocket.last!.actions()).toContain("undo");
  });

  it("disables the Undo when the turn's budget is spent", async () => {
    const c = await mountGame(declareAttackers({ undos_remaining: 0 }));
    const undo = buttons(dockOf(c), /undo/i)[0]!;
    expect(undo.disabled).toBe(true);
    expect(undo.getAttribute("aria-label")).toContain("no undos remaining");
  });

  it("has no Undo left in the ⋯ menu", async () => {
    const c = await mountGame(declareAttackers());
    click(c.querySelector<HTMLElement>('button[aria-label="more actions"]')!);
    flushSync();
    const items = [...c.querySelectorAll<HTMLElement>('[role="menuitem"]')];
    expect(items.length).toBeGreaterThan(0);
    expect(items.some((i) => /undo/i.test(i.textContent ?? ""))).toBe(false);
  });

  it("makes No blocks the dock's primary for a defender who owes blocks, and Done blocking once staged", async () => {
    const c = await mountGame(declareBlockers(false));
    const dock = dockOf(c);
    const dlg = dock.querySelector('[role="dialog"][aria-label="declare blockers"]')!;
    expect(dlg).not.toBeNull();
    const noBlocks = buttons(dlg, "No blocks")[0]!;
    expect(noBlocks.classList.contains("primary")).toBe(true);
    // It takes the bar: next and Pass turn give way.
    expect(buttons(c, "next")).toHaveLength(0);
    expect(buttons(stripOf(c), /block/i)).toHaveLength(0);
    click(noBlocks);
    expect(FakeSocket.last!.actions()).toContain("finish_blocks");

    snapshot(declareBlockers(true));
    expect(buttons(dockOf(c), "Done blocking")).toHaveLength(1);
    expect(buttons(c, "No blocks")).toHaveLength(0);
  });

  it("shows a combat selection's hint and Cancel in the dock; Cancel and Escape clear it", async () => {
    const c = await mountGame(declareAttackers());
    const select = c.querySelector<HTMLElement>('[data-testid="select-b1"]')!;
    click(select);
    flushSync();
    const dock = dockOf(c);
    let dlg = dock.querySelector('[role="dialog"][aria-label="attacking with Grizzly Bears"]');
    expect(dlg).not.toBeNull();
    expect(dlg!.textContent).toContain("click an opponent's seat to commit");
    // The selection outranks the attack row; the strip has no hint.
    expect(dock.querySelector('[aria-label="declare attackers"]')).toBeNull();
    expect(stripOf(c).querySelector(".combat-hint")).toBeNull();
    click(buttons(dlg!, "Cancel")[0]!);
    flushSync();
    expect(dock.querySelector('[aria-label="attacking with Grizzly Bears"]')).toBeNull();
    expect(dock.querySelector('[role="dialog"][aria-label="declare attackers"]')).not.toBeNull();

    click(select);
    flushSync();
    dlg = dock.querySelector('[aria-label="attacking with Grizzly Bears"]');
    expect(dlg).not.toBeNull();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(dock.querySelector('[aria-label="attacking with Grizzly Bears"]')).toBeNull();
  });

  it("answers an attack-tax refusal of attack-with-all in the dock, not the strip", async () => {
    const c = await mountGame(declareAttackers());
    click(buttons(dockOf(c), "Attack Opp with all 2 creatures")[0]!);
    const frame = FakeSocket.last!.sent.filter((f) => f.kind === "action").at(-1) as unknown as {
      id: string;
    };
    FakeSocket.last!.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "error",
        id: frame.id,
        payload: {
          code: "attack_tax_unpaid",
          message: "can't pay",
          reason: "{4}",
          missing: ["{2}"],
        },
      }),
    });
    flushSync();
    const alert = dockOf(c).querySelector('[role="dialog"] [role="alert"]')!;
    expect(alert).not.toBeNull();
    expect(alert.textContent).toContain("costs {4}");
    expect(buttons(alert, "Choose attackers…")).toHaveLength(1);
    // Not drawn twice: the strip shows no toast for it.
    expect(stripOf(c).querySelector('[role="alert"]')).toBeNull();
  });
});
