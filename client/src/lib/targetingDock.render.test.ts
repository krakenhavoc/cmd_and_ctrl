// @vitest-environment jsdom
//
// targetingDock.render.test.ts — ADR 0111 Delivery PR 4 (S56, #1958),
// through the real Game route. Targeting's Done and Cancel and the
// insufficient-mana prompt left the attention strip for the action
// dock; one Enter / Escape handler serves every request; the strip is
// `region "attention"`; and No blocks, when the defender holds
// priority, declares no blocks and then passes (owner decision,
// 2026-10-02). The board is a stub that renders the strip Game.svelte
// hands it, so these read what is (and is not) in it.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { get } from "svelte/store";

import { PROTOCOL_VERSION, type CardView, type GameView, type PlayerView } from "./protocol";
import { session, type Session } from "./session";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals, pushModalLayer } from "./modalLayers";
import { begin, setConfirmHandler, targeting, togglePick } from "./targeting";
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

interface SentFrame {
  kind: string;
  id: string;
  payload?: { type?: string; params?: Record<string, unknown> };
}

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  sent: SentFrame[] = [];
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
  actionFrames(): SentFrame[] {
    return this.sent.filter((f) => f.kind === "action");
  }
  actions(): string[] {
    return this.actionFrames().map((f) => f.payload?.type ?? "");
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

const elves = {
  instance_id: "elves",
  name: "Llanowar Elves",
  owner: ME,
  controller: ME,
  type_line: "Creature — Elf Druid",
} as CardView;

function mainPhase(): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      seat(ME, "Me", 0, { hand: zone("hand", ME, [elves]) } as Partial<PlayerView>),
      seat(OPP, "Opp", 1),
    ],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 2,
      number: 2,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
  } as unknown as GameView;
}

// Three seats: Opp attacks both Me and Opp2, so the viewer is one of
// two defenders. Me holds priority (the active player passed).
function blockersFor(opts: {
  staged: boolean;
  pending: number[];
  declared?: number[];
  holder: number;
  step?: string;
}): GameView {
  return {
    ...mainPhase(),
    seats: [seat(ME, "Me", 0), seat(OPP, "Opp", 1), seat(OPP2, "Opp2", 2)],
    battlefield: zone("battlefield", undefined, [
      creature("a1", OPP, { attacking_target: ME } as Partial<CardView>),
      creature("a2", OPP, { attacking_target: OPP2 } as Partial<CardView>),
      creature("b1", ME, opts.staged ? ({ blocking_target: "a1" } as Partial<CardView>) : {}),
    ]),
    turn: {
      seq: 5,
      number: 3,
      active_seat: 1,
      priority_holder: opts.holder,
      phase: "combat",
      step: opts.step ?? "declare_blockers",
      block_pending_seats: opts.pending,
      block_decision_seats: opts.pending,
      blocks_declared_seats: opts.declared ?? [],
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
  targeting.set(null);
  setConfirmHandler(null);
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  // A keydown arms the audio (sounds.ts preloads); jsdom has no load().
  vi.spyOn(HTMLMediaElement.prototype, "load").mockImplementation(() => {});
  session.set(playerSession());
});

afterEach(() => {
  cleanup();
  targeting.set(null);
  setConfirmHandler(null);
  resetModals();
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

function errorFor(frameID: string, payload: Record<string, unknown>): void {
  FakeSocket.last!.emit("message", {
    data: JSON.stringify({ v: PROTOCOL_VERSION, kind: "error", id: frameID, payload }),
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
const stripOf = (c: ParentNode) =>
  c.querySelector<HTMLElement>('[role="region"][aria-label="attention"]')!;
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
function keydown(key: string, target: EventTarget = window): void {
  target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true }));
  flushSync();
}

const arc = (): CardView =>
  ({
    instance_id: "arc",
    name: "Arc Lightning",
    owner: ME,
    controller: ME,
    legal_targets: { min: 1, max: 3, players: [OPP] },
  }) as CardView;

describe("the attention strip", () => {
  it('is region "attention" (ADR 0076 step 9\'s anchor)', async () => {
    const c = await mountGame(mainPhase());
    const strip = stripOf(c);
    expect(strip).not.toBeNull();
    expect(strip.getAttribute("role")).toBe("region");
  });
});

describe("targeting in the dock (ADR 0111 §6)", () => {
  it("puts the prompt, Done and Cancel in the dock, and nothing of it in the strip", async () => {
    const c = await mountGame(mainPhase());
    // With priority and nothing asked, the status line says what next
    // will do…
    expect(dockOf(c).querySelector(".pass-hint")).not.toBeNull();
    begin(arc(), "any");
    flushSync();
    const dock = dockOf(c);
    // …and not while a request has taken the bar and next is gone.
    expect(dock.querySelector(".pass-hint")).toBeNull();
    const dlg = dock.querySelector(
      '[role="dialog"][aria-label="Select target for Arc Lightning"]',
    )!;
    expect(dlg).not.toBeNull();
    expect(dlg.textContent).toContain("Click a player or creature to target Arc Lightning");
    expect(dlg.textContent).toContain("0/3 picked");
    const done = buttons(dlg, "Done")[0]!;
    expect(done.classList.contains("primary")).toBe(true);
    expect(done.disabled).toBe(true);
    expect(buttons(dlg, "Cancel")).toHaveLength(1);
    // It takes the bar: next and Pass turn give way.
    expect(buttons(c, "next")).toHaveLength(0);

    // Nothing in the strip asks about it any more, and there is one of
    // each on the page.
    const strip = stripOf(c);
    expect(strip.querySelector('[role="dialog"]')).toBeNull();
    expect(buttons(strip, /done|cancel/i)).toHaveLength(0);
    expect(buttons(c, "Done")).toHaveLength(1);
    expect(c.querySelectorAll('[aria-label="Select target for Arc Lightning"]')).toHaveLength(1);

    click(buttons(dlg, "Cancel")[0]!);
    flushSync();
    expect(get(targeting)).toBeNull();
    expect(dock.querySelector('[role="dialog"]')).toBeNull();
    expect(buttons(c, "next")).toHaveLength(1);
  });

  it("Done confirms once enough are picked", async () => {
    const c = await mountGame(mainPhase());
    const confirmed = vi.fn();
    setConfirmHandler(confirmed);
    begin(arc(), "any");
    targeting.set(togglePick(get(targeting)!, { kind: "player", id: OPP }));
    flushSync();
    const done = buttons(dockOf(c), "Done")[0]!;
    expect(done.disabled).toBe(false);
    expect(dockOf(c).textContent).toContain("1/3 picked");
    click(done);
    expect(confirmed).toHaveBeenCalledTimes(1);
  });
});

describe("the one Enter / Escape handler", () => {
  it("Enter presses targeting's Done and Escape its Cancel", async () => {
    await mountGame(mainPhase());
    const confirmed = vi.fn();
    setConfirmHandler(confirmed);
    begin(arc(), "any");
    flushSync();
    // Not enough picked: Done is disabled, so Enter does nothing.
    keydown("Enter");
    expect(confirmed).not.toHaveBeenCalled();
    targeting.set(togglePick(get(targeting)!, { kind: "player", id: OPP }));
    flushSync();
    keydown("Enter");
    expect(confirmed).toHaveBeenCalledTimes(1);
    keydown("Escape");
    expect(get(targeting)).toBeNull();
  });

  // #1659 (was targetingModalKeys.render.test.ts): a modal opened during
  // the same cast (the mode picker, an X prompt, a cost picker,
  // DivideDamageModal, ChoicePromptModal) owns Escape and Enter. Escape
  // must not also cancel the walk underneath it, nor Enter confirm it.
  it("stands down behind a modal layer, and once it closes works again (#1659)", async () => {
    await mountGame(mainPhase());
    const confirmed = vi.fn();
    setConfirmHandler(confirmed);
    const unregister = pushModalLayer();
    begin(arc(), "any");
    targeting.set(togglePick(get(targeting)!, { kind: "player", id: OPP }));
    flushSync();
    keydown("Enter");
    keydown("Escape");
    expect(confirmed).not.toHaveBeenCalled();
    expect(get(targeting)).not.toBeNull();
    unregister();
    flushSync();
    keydown("Enter");
    expect(confirmed).toHaveBeenCalledTimes(1);
  });

  it("is skipped while an input has focus", async () => {
    const c = await mountGame(mainPhase());
    const confirmed = vi.fn();
    setConfirmHandler(confirmed);
    begin(arc(), "any");
    targeting.set(togglePick(get(targeting)!, { kind: "player", id: OPP }));
    flushSync();
    const input = document.createElement("input");
    c.appendChild(input);
    input.focus();
    keydown("Enter", input);
    keydown("Escape", input);
    expect(confirmed).not.toHaveBeenCalled();
    expect(get(targeting)).not.toBeNull();
  });

  it("never presses next: Enter and Escape with no request send nothing", async () => {
    const c = await mountGame(mainPhase());
    expect(buttons(dockOf(c), "next")[0]!.disabled).toBe(false);
    const before = FakeSocket.last!.actions().length;
    keydown("Enter");
    keydown("Escape");
    keydown("Enter", document.body);
    expect(FakeSocket.last!.actions().slice(before)).not.toContain("pass_priority");
    expect(FakeSocket.last!.actions().length).toBe(before);
  });

  it("Escape cancels a combat selection (absorbed from Game.svelte's window handler)", async () => {
    const view = {
      ...mainPhase(),
      battlefield: zone("battlefield", undefined, [creature("b1", ME)]),
      turn: {
        seq: 3,
        number: 3,
        active_seat: 0,
        priority_holder: 0,
        phase: "combat",
        step: "declare_attackers",
      },
    } as unknown as GameView;
    const c = await mountGame(view);
    click(c.querySelector<HTMLElement>('[data-testid="select-b1"]')!);
    flushSync();
    expect(dockOf(c).querySelector('[aria-label="attacking with Grizzly Bears"]')).not.toBeNull();
    keydown("Enter");
    expect(dockOf(c).querySelector('[aria-label="attacking with Grizzly Bears"]')).not.toBeNull();
    keydown("Escape");
    expect(dockOf(c).querySelector('[aria-label="attacking with Grizzly Bears"]')).toBeNull();
  });
});

describe("insufficient mana in the dock (S15, ADR 0111 PR 4)", () => {
  async function refusedCast(): Promise<HTMLElement> {
    const c = await mountGame(mainPhase());
    FakeSocket.last!.sent.length = 0;
    // A cast the server refused for mana: the dock offers the way on.
    errorFor("cast-1", {
      code: "insufficient_mana",
      message: "not enough mana",
      missing: ["{G}"],
      card_id: "elves",
    });
    return c;
  }

  it("is a dock request with Auto-tap & cast and Cast anyway, not a strip toast", async () => {
    const c = await refusedCast();
    const dlg = dockOf(c).querySelector('[role="dialog"][aria-label="insufficient mana"]')!;
    expect(dlg).not.toBeNull();
    expect(dlg.textContent).toContain("Insufficient mana for Llanowar Elves");
    expect(dlg.textContent).toContain("missing {G}");
    const auto = buttons(dlg, "Auto-tap & cast")[0]!;
    expect(auto.classList.contains("primary")).toBe(true);
    expect(buttons(dlg, "Cast anyway")).toHaveLength(1);
    expect(buttons(dlg, "Cancel")).toHaveLength(1);
    // Nothing about it in the strip: no toast, no rejection.
    expect(stripOf(c).querySelector('[role="alert"]')).toBeNull();
    expect(buttons(c, "Auto-tap & cast")).toHaveLength(1);
    expect(buttons(c, "Cast anyway")).toHaveLength(1);
  });

  it("Cast anyway re-sends the cast with force_cast", async () => {
    const c = await refusedCast();
    click(buttons(dockOf(c), "Cast anyway")[0]!);
    flushSync();
    const cast = FakeSocket.last!.actionFrames().find((f) => f.payload?.type === "cast_spell")!;
    expect(cast).toBeTruthy();
    expect(cast.payload?.params).toMatchObject({ instance_id: "elves", force_cast: true });
    expect(dockOf(c).querySelector('[aria-label="insufficient mana"]')).toBeNull();
  });

  it("Auto-tap & cast opens the preview and closes the request", async () => {
    const c = await refusedCast();
    click(buttons(dockOf(c), "Auto-tap & cast")[0]!);
    flushSync();
    expect(dockOf(c).querySelector('[aria-label="insufficient mana"]')).toBeNull();
    // The preview is still a centred modal until Delivery PR 6.
    expect(c.querySelector('[role="dialog"][aria-modal="true"]')).not.toBeNull();
  });

  it("Escape cancels it; Enter presses Auto-tap & cast", async () => {
    let c = await refusedCast();
    keydown("Escape");
    expect(dockOf(c).querySelector('[aria-label="insufficient mana"]')).toBeNull();
    cleanup();

    c = await refusedCast();
    keydown("Enter");
    expect(dockOf(c).querySelector('[aria-label="insufficient mana"]')).toBeNull();
    expect(c.querySelector('[role="dialog"][aria-modal="true"]')).not.toBeNull();
  });
});

describe("No blocks, one click (owner decision 2026-10-02)", () => {
  it("declares no blocks, then passes once the server has accepted it", async () => {
    const c = await mountGame(blockersFor({ staged: false, pending: [0, 2], holder: 0 }));
    const noBlocks = buttons(dockOf(c), "No blocks")[0]!;
    click(noBlocks);
    // Only the declaration so far: the pass waits for the server.
    expect(FakeSocket.last!.actions()).toEqual(["finish_blocks"]);

    // Another seat's frame that still has the declaration open: wait.
    snapshot(blockersFor({ staged: false, pending: [0, 2], holder: 0 }));
    expect(FakeSocket.last!.actions()).toEqual(["finish_blocks"]);

    // Accepted: the viewer's declaration is done, Opp2's is not, and the
    // viewer still holds priority. Now the pass, once.
    snapshot(blockersFor({ staged: false, pending: [2], declared: [0], holder: 0 }));
    expect(FakeSocket.last!.actions()).toEqual(["finish_blocks", "pass_priority"]);
    snapshot(blockersFor({ staged: false, pending: [2], declared: [0], holder: 0 }));
    expect(FakeSocket.last!.actions().filter((a) => a === "pass_priority")).toHaveLength(1);
  });

  it("does not pass when the finish hands priority to the active player", async () => {
    const c = await mountGame(blockersFor({ staged: false, pending: [0], holder: 0 }));
    click(buttons(dockOf(c), "No blocks")[0]!);
    // The last declaration: CR 509.2, priority goes to the active player.
    snapshot(blockersFor({ staged: false, pending: [], declared: [0, 2], holder: 1 }));
    expect(FakeSocket.last!.actions()).toEqual(["finish_blocks"]);
  });

  it("does not pass when the server refuses the declaration", async () => {
    const c = await mountGame(blockersFor({ staged: false, pending: [0, 2], holder: 0 }));
    click(buttons(dockOf(c), "No blocks")[0]!);
    const finish = FakeSocket.last!.actionFrames()[0]!;
    errorFor(finish.id, {
      code: "illegal_block",
      message: "must block",
      reason: "block_requirement",
    });
    snapshot(blockersFor({ staged: false, pending: [2], declared: [0], holder: 0 }));
    expect(FakeSocket.last!.actions()).toEqual(["finish_blocks"]);
  });

  it("only declares when the defender does not hold priority", async () => {
    const c = await mountGame(blockersFor({ staged: false, pending: [0, 2], holder: 1 }));
    click(buttons(dockOf(c), "No blocks")[0]!);
    snapshot(blockersFor({ staged: false, pending: [2], declared: [0], holder: 0 }));
    // Priority reaching the viewer later is a window of its own.
    expect(FakeSocket.last!.actions()).toEqual(["finish_blocks"]);
  });

  it("Done blocking is a separate confirm and does not pass", async () => {
    const c = await mountGame(blockersFor({ staged: true, pending: [0, 2], holder: 0 }));
    const done = buttons(dockOf(c), "Done blocking")[0]!;
    expect(done).toBeTruthy();
    click(done);
    snapshot(blockersFor({ staged: true, pending: [2], declared: [0], holder: 0 }));
    expect(FakeSocket.last!.actions()).toEqual(["finish_blocks"]);
    // next is back, for the player to pass when they choose.
    expect(buttons(dockOf(c), "next")).toHaveLength(1);
  });
});
