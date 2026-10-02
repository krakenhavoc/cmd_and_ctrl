// @vitest-environment jsdom
//
// choiceDock.render.test.ts — ADR 0111 Delivery PR 5 (S56, #1958),
// through the real Game route. The small pending choices left
// ChoicePromptModal for the action dock: the yes/no family, pay_unless
// without picks, coin_call, loop_shortcut, mana_pick, choose_color and
// a short option_pick / entry_controller. So did the open vote (out of
// VotingPanel) and the game-over Back to lobby (out of the strip).
//
// For each: it is drawn in the dock, inside one non-modal dialog named
// as the modal was; its buttons keep their names; it sends the right
// answer; nothing of it is drawn in a modal; and Enter / Escape do not
// answer it (ADR 0111 §1: a yes/no does not take Enter), while the keys
// it had (Y / N, H / T / S, 1-9) still do. The board is a stub that
// renders the strip Game.svelte hands it.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import {
  PROTOCOL_VERSION,
  type GameView,
  type PendingChoiceView,
  type PlayerView,
} from "./protocol";
import { session, type Session } from "./session";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { targeting, setConfirmHandler } from "./targeting";
import { defaultSettings, settings } from "./settings";
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
}

const ME = "p-1";
const OPP = "p-2";

const zone = (kind: string, owner?: string, cards: unknown[] = []) => ({
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

// The viewer holds priority on their own main phase, so `next` is live
// until a request takes the bar.
function table(over: Partial<GameView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Opp", 1)],
    battlefield: zone("battlefield", undefined, [
      {
        instance_id: "sage",
        name: "Reclamation Sage",
        controller: ME,
        owner: ME,
        type_line: "Creature — Elf Shaman",
      },
    ]),
    stack: zone("stack"),
    exile: zone("exile", undefined, [
      { instance_id: "bolt", name: "Lightning Bolt", owner: ME, controller: ME },
    ]),
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
    ...over,
  } as unknown as GameView;
}

function choice(over: Partial<PendingChoiceView>): PendingChoiceView {
  return {
    id: "choice-1",
    kind: "trigger_prompt",
    chooser: ME,
    from_player: ME,
    count: 0,
    options: [],
    ...over,
  } as PendingChoiceView;
}

const withChoice = (c: Partial<PendingChoiceView>, over: Partial<GameView> = {}): GameView =>
  table({ pending_choices: [choice(c)], ...over });

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
  settings.set(defaultSettings());
  resetDock();
  resetModals();
  targeting.set(null);
  setConfirmHandler(null);
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  vi.spyOn(HTMLMediaElement.prototype, "load").mockImplementation(() => {});
  session.set(playerSession());
});

afterEach(() => {
  cleanup();
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
const dialogNamed = (root: ParentNode, name: string): HTMLElement | null =>
  [...root.querySelectorAll<HTMLElement>('[role="dialog"]')].find(
    (d) => d.getAttribute("aria-label") === name,
  ) ?? null;
function keydown(key: string, target: EventTarget = window): void {
  target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true }));
  flushSync();
}
const answers = () =>
  FakeSocket.last!.actionFrames()
    .filter((f) => f.payload?.type === "resolve_choice")
    .map((f) => f.payload?.params);

// The prompt is in the dock, in one dialog of that name, and nowhere
// else: no backdrop, no modal heading, not in the strip.
function expectInDock(c: HTMLElement, name: string): HTMLElement {
  const dock = dockOf(c);
  const dlg = dialogNamed(dock, name);
  expect(dlg, `dialog "${name}" in the dock`).not.toBeNull();
  expect(dlg!.getAttribute("aria-modal")).toBeNull();
  expect(dlg!.dataset.rank).toBe("choice");
  expect(
    [...c.querySelectorAll("[aria-label]")].filter((e) => e.getAttribute("aria-label") === name),
  ).toHaveLength(1);
  expect(c.querySelector(".prompt-backdrop")).toBeNull();
  expect(c.querySelector("#choice-title")).toBeNull();
  expect(stripOf(c).querySelector('[role="dialog"]')).toBeNull();
  return dlg!;
}

describe("the yes/no family, inline in the dock", () => {
  const cases: {
    kind: string;
    over: Partial<PendingChoiceView>;
    name: string;
    yes: string;
    no: string;
  }[] = [
    {
      kind: "trigger_prompt",
      over: {
        reason: "Reclamation Sage — destroy target artifact or enchantment?",
        source: "sage",
      },
      name: "Reclamation Sage — destroy target artifact or enchantment?",
      yes: "Yes",
      no: "No",
    },
    {
      kind: "optional_replacement",
      over: { reason: "Put your commander into the command zone?" },
      name: "Put your commander into the command zone?",
      yes: "Yes",
      no: "No",
    },
    {
      kind: "confirm",
      over: {
        reason: "Sylvan Library — pay 4 life for this card?",
        accept_label: "Pay 4 life",
        decline_label: "Put it back",
      },
      name: "Sylvan Library — pay 4 life for this card?",
      yes: "Pay 4 life",
      no: "Put it back",
    },
    {
      kind: "may_cast",
      over: {
        reason: "Cascade — cast Lightning Bolt free?",
        may_cast_keyword: "cascade",
        may_cast_card: "bolt",
      },
      name: "Cascade — cast Lightning Bolt free?",
      yes: "Cast it free",
      no: "Put it on the bottom",
    },
    {
      kind: "entry_pay_life",
      over: { reason: "Hallowed Fountain — pay 2 life?", pay_cost: "2 life" },
      name: "Hallowed Fountain — pay 2 life?",
      yes: "Pay 2 life",
      no: "Enter tapped",
    },
    {
      kind: "pay_unless",
      over: { reason: "Smothering Tithe — pay {2}?", pay_cost: "{2}" },
      name: "Smothering Tithe — pay {2}?",
      yes: "Pay {2}",
      no: "Don't pay",
    },
  ];

  for (const tc of cases) {
    it(`${tc.kind}: drawn in the dock as "${tc.name}", and answers apply true / false`, async () => {
      const c = await mountGame(withChoice({ kind: tc.kind, ...tc.over }));
      const dlg = expectInDock(c, tc.name);
      expect(dlg.textContent).toContain(tc.name);

      // It takes the bar: the answer in the corner, next and Pass turn
      // gone until it closes. One of each button on the page.
      const yes = buttons(dlg, tc.yes);
      const no = buttons(dlg, tc.no);
      expect(yes).toHaveLength(1);
      expect(no).toHaveLength(1);
      expect(buttons(c, tc.yes)).toHaveLength(1);
      expect(yes[0].classList.contains("primary")).toBe(true);
      expect(yes[0].getAttribute("aria-keyshortcuts")).toBe("Y");
      expect(no[0].getAttribute("aria-keyshortcuts")).toBe("N");
      expect(buttons(c, "next")).toHaveLength(0);
      expect(buttons(c, "Pass turn")).toHaveLength(0);

      click(no[0]);
      click(yes[0]);
      expect(answers()).toEqual([
        { choice_id: "choice-1", apply: false },
        { choice_id: "choice-1", apply: true },
      ]);

      // The server drains it: the dialog goes, next comes back.
      snapshot(table());
      expect(dialogNamed(dockOf(c), tc.name)).toBeNull();
      expect(buttons(c, "next")).toHaveLength(1);
    });
  }

  it("a doubled trigger keeps its note in the dialog's name, as the heading had it", async () => {
    const c = await mountGame(
      withChoice({
        reason: "Reclamation Sage — destroy target artifact or enchantment?",
        doubled_by: "card-x",
        doubled_by_name: "Panharmonicon",
      }),
    );
    const dlg = dockOf(c).querySelector<HTMLElement>('[role="dialog"]')!;
    expect(dlg.getAttribute("aria-label")).toMatch(
      /^Reclamation Sage — destroy target artifact or enchantment\? additional \(Panharmonicon\)$/,
    );
  });

  it("warns when a may trigger has no legal target", async () => {
    const c = await mountGame(
      withChoice({ reason: "Reclamation Sage — destroy?", no_legal_target: true }),
    );
    const hint = expectInDock(c, "Reclamation Sage — destroy?").querySelector(".dock-hint")!;
    expect(hint.classList.contains("warn")).toBe(true);
    expect(hint.textContent).toContain("No legal target");
  });

  it("is not drawn for a seat that is not the chooser", async () => {
    const c = await mountGame(withChoice({ reason: "Reclamation Sage — destroy?", chooser: OPP }));
    expect(dockOf(c).querySelector('[role="dialog"]')).toBeNull();
    expect(buttons(c, "Yes")).toHaveLength(0);
    expect(buttons(c, "next")).toHaveLength(1);
  });

  it("pay_unless with card picks is a sheet in the dock, not inline (PR 6)", async () => {
    const c = await mountGame(
      withChoice({
        kind: "pay_unless",
        reason: "Echo — discard a card?",
        pay_cost: "Discard a card",
        pay_cards: { action: "discard", count: 1, options: [] },
      }),
    );
    const dlg = dialogNamed(dockOf(c), "Echo — discard a card?");
    expect(dlg).not.toBeNull();
    expect(dlg!.querySelector(".dock-sheet")).not.toBeNull();
    expect(c.querySelector(".prompt-backdrop")).toBeNull();
  });
});

describe("keys for the yes/no family (ADR 0111 §1)", () => {
  const REASON = "Reclamation Sage — destroy target artifact or enchantment?";

  it("Enter and Escape do not answer it; Y and N do", async () => {
    const c = await mountGame(withChoice({ reason: REASON }));
    expectInDock(c, REASON);

    keydown("Enter");
    keydown("Escape");
    expect(answers()).toEqual([]);

    keydown("n");
    expect(answers()).toEqual([{ choice_id: "choice-1", apply: false }]);
    keydown("Y");
    expect(answers()).toEqual([
      { choice_id: "choice-1", apply: false },
      { choice_id: "choice-1", apply: true },
    ]);
  });

  it("moves focus to the question, not to Yes, so a stray Enter accepts nothing", async () => {
    const c = await mountGame(table());
    (document.activeElement as HTMLElement | null)?.blur?.();
    snapshot(withChoice({ reason: REASON }));
    const dlg = expectInDock(c, REASON);
    expect(document.activeElement).toBe(dlg);
    // Enter with focus on the question presses nothing.
    keydown("Enter", dlg);
    expect(answers()).toEqual([]);
  });

  it("the global shortcuts stand down while it is open (it owns Y / N)", async () => {
    const c = await mountGame(withChoice({ reason: REASON }));
    expectInDock(c, REASON);
    // Space is pass priority; it must not pass under an open question.
    keydown(" ");
    expect(FakeSocket.last!.actionFrames().map((f) => f.payload?.type)).not.toContain(
      "pass_priority",
    );
  });
});

describe("coin_call, inline", () => {
  it("Heads, Tails and Stop answer it, by click and by H / T / S; Enter does not", async () => {
    const c = await mountGame(
      withChoice({
        kind: "coin_call",
        reason: "Fiery Gambit — call the flip",
        coins: 1,
        wins: 2,
        allow_stop: true,
      }),
    );
    const dlg = expectInDock(c, "Fiery Gambit — call the flip");
    expect(dlg.textContent).toContain("You have won 2 flips so far");
    expect(buttons(dlg, "Tails")[0].classList.contains("primary")).toBe(true);
    click(buttons(dlg, "Heads")[0]);
    click(buttons(dlg, "Stop")[0]);
    keydown("Enter");
    keydown("Escape");
    keydown("t");
    expect(answers()).toEqual([
      { choice_id: "choice-1", call: "heads" },
      { choice_id: "choice-1", call: "stop" },
      { choice_id: "choice-1", call: "tails" },
    ]);
  });

  it("has no Stop unless the effect allows one", async () => {
    const c = await mountGame(withChoice({ kind: "coin_call", reason: "Call it" }));
    const dlg = expectInDock(c, "Call it");
    expect(buttons(dlg, "Stop")).toHaveLength(0);
    keydown("s");
    expect(answers()).toEqual([]);
  });
});

describe("loop_shortcut, inline (CR 732)", () => {
  it("Stop here sends 0; Resolve N more sends the number in the field", async () => {
    const c = await mountGame(
      withChoice({ kind: "loop_shortcut", reason: "Loop detected", loop_count: 4 }),
    );
    const dlg = expectInDock(c, "Loop detected");
    const field = dlg.querySelector<HTMLInputElement>('input[type="number"]')!;
    expect(field).not.toBeNull();
    expect(buttons(dlg, "Resolve 10 more")).toHaveLength(1);

    click(buttons(dlg, "Stop here")[0]);
    field.value = "25";
    field.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    click(buttons(dlg, "Resolve 25 more")[0]);
    expect(answers()).toEqual([
      { choice_id: "choice-1", iterations: 0 },
      { choice_id: "choice-1", iterations: 25 },
    ]);
  });

  it("Enter in the field sends the typed number; Enter and Escape elsewhere do not", async () => {
    const c = await mountGame(withChoice({ kind: "loop_shortcut", reason: "Loop detected" }));
    const dlg = expectInDock(c, "Loop detected");
    keydown("Enter");
    keydown("Escape");
    expect(answers()).toEqual([]);
    const field = dlg.querySelector<HTMLInputElement>('input[type="number"]')!;
    field.value = "3";
    field.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    keydown("Enter", field);
    expect(answers()).toEqual([{ choice_id: "choice-1", iterations: 3 }]);
  });
});

describe("mana_pick and choose_color, inline", () => {
  it("mana_pick: the symbols in the dock, answered by click or 1-9, no bar", async () => {
    const c = await mountGame(
      withChoice({
        kind: "mana_pick",
        reason: "Birds of Paradise — pick a color",
        color_options: ["G", "U", "W"],
      }),
    );
    const dlg = expectInDock(c, "Birds of Paradise — pick a color");
    const symbols = dlg.querySelectorAll<HTMLButtonElement>(".mana-option");
    expect(symbols).toHaveLength(3);
    // No action bar: a symbol is the answer. And not next either.
    expect(dlg.querySelector(".dock-bar")).toBeNull();
    expect(buttons(c, "next")).toHaveLength(0);
    click(symbols[2]);
    keydown("2");
    keydown("Escape");
    expect(answers()).toEqual([
      { choice_id: "choice-1", color: "W" },
      { choice_id: "choice-1", color: "U" },
    ]);
  });

  it("choose_color: says what the colour is for, and answers with it", async () => {
    const c = await mountGame(
      withChoice({
        kind: "choose_color",
        reason: "Coldsteel Heart — choose a color",
        color_options: ["W", "U", "B", "R", "G"],
        color_purpose: "mana",
      }),
    );
    const dlg = expectInDock(c, "Coldsteel Heart — choose a color");
    expect(dlg.querySelector(".dock-hint")?.textContent?.length).toBeGreaterThan(0);
    click(dlg.querySelectorAll<HTMLButtonElement>(".mana-option")[3]);
    expect(answers()).toEqual([{ choice_id: "choice-1", color: "R" }]);
  });
});

describe("option_pick and entry_controller, inline when short", () => {
  it("option_pick: one button per option, answered with its index", async () => {
    const c = await mountGame(
      withChoice({
        kind: "option_pick",
        reason: "Torment of Hailfire — choose one",
        pick_options: [{ label: "Lose 3 life" }, { label: "Discard a card" }],
      }),
    );
    const dlg = expectInDock(c, "Torment of Hailfire — choose one");
    expect(dlg.querySelector(".dock-row.stack")).not.toBeNull();
    keydown("Enter");
    keydown("Escape");
    click(buttons(dlg, "Discard a card")[0]);
    expect(answers()).toEqual([{ choice_id: "choice-1", option_index: 1 }]);
  });

  it("entry_controller: one button per opponent", async () => {
    const c = await mountGame(
      withChoice({
        kind: "entry_controller",
        reason: "Captive Audience — choose an opponent",
        pick_options: [{ label: "Opp", player: OPP }],
      }),
    );
    const dlg = expectInDock(c, "Captive Audience — choose an opponent");
    click(buttons(dlg, "Opp")[0]);
    expect(answers()).toEqual([{ choice_id: "choice-1", option_index: 0 }]);
  });
});

describe("a refusal of the answer (#624)", () => {
  it("is announced in the dock's prompt area as Not accepted", async () => {
    const c = await mountGame(
      withChoice({ kind: "pay_unless", reason: "Rhystic Study — pay {1}?", pay_cost: "{1}" }),
    );
    const dlg = expectInDock(c, "Rhystic Study — pay {1}?");
    click(buttons(dlg, "Pay {1}")[0]);
    const frame = FakeSocket.last!.actionFrames().at(-1)!;
    FakeSocket.last!.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "error",
        id: frame.id,
        payload: { code: "cannot_pay", message: "not enough mana in pool" },
      }),
    });
    flushSync();
    const alert = dlg.querySelector('[role="alert"]');
    expect(alert?.textContent).toContain("Not accepted");
    expect(alert?.textContent).toContain("not enough mana in pool");
    // Still answerable.
    expect(buttons(dlg, "Don't pay")).toHaveLength(1);
    // Announced once: the strip's rejection toast stands down for the
    // error the dock is showing.
    expect(stripOf(c).querySelector('[role="alert"]')).toBeNull();
    expect(c.querySelectorAll('[role="alert"]')).toHaveLength(1);
  });

  it("leaves any other error to the strip", async () => {
    const c = await mountGame(
      withChoice({ kind: "pay_unless", reason: "Rhystic Study — pay {1}?", pay_cost: "{1}" }),
    );
    FakeSocket.last!.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "error",
        id: "some-other-frame",
        payload: { code: "illegal_action", message: "not your priority" },
      }),
    });
    flushSync();
    expect(stripOf(c).querySelector('[role="alert"]')?.textContent).toContain("not your priority");
    expect(dockOf(c).querySelector('[role="alert"]')).toBeNull();
  });
});

describe("the vote, in the dock", () => {
  const vote = {
    id: "v1",
    topic: "Monarchy?",
    options: ["yes", "no"],
    initiator: OPP,
    ballots: { [OPP]: 0 } as Record<string, number>,
  };

  it("draws the options with their tally, keeps next, and casts a ballot", async () => {
    const c = await mountGame(table({ vote }));
    const dlg = dialogNamed(dockOf(c), "open vote")!;
    expect(dlg).not.toBeNull();
    expect(dlg.textContent).toContain("Monarchy?");
    expect(dlg.textContent).toContain("called by Opp");
    // A vote never stops the game: next and Pass turn stay.
    expect(buttons(c, "next")).toHaveLength(1);
    expect(buttons(c, "Pass turn")).toHaveLength(1);
    // No floating panel as well.
    expect(c.querySelector(".vote-modal")).toBeNull();

    const yes = buttons(dlg, "yes 1")[0];
    expect(yes).toBeTruthy();
    click(buttons(dlg, "no 0")[0]);
    click(buttons(dlg, "end vote")[0]);
    const sent = FakeSocket.last!.actionFrames().map((f) => [f.payload?.type, f.payload?.params]);
    expect(sent).toEqual([
      ["cast_vote", { option: 1 }],
      ["end_vote", undefined],
    ]);
  });

  it("marks the viewer's own ballot pressed", async () => {
    const c = await mountGame(table({ vote: { ...vote, ballots: { [ME]: 1 } } }));
    const dlg = dialogNamed(dockOf(c), "open vote")!;
    expect(buttons(dlg, "no 1")[0].getAttribute("aria-pressed")).toBe("true");
    expect(buttons(dlg, "yes 0")[0].getAttribute("aria-pressed")).toBe("false");
  });

  it("gives way to a pending choice, and comes back when it is answered", async () => {
    const c = await mountGame(table({ vote }));
    snapshot(withChoice({ reason: "Reclamation Sage — destroy?" }, { vote }));
    expect(dialogNamed(dockOf(c), "open vote")).toBeNull();
    expectInDock(c, "Reclamation Sage — destroy?");
    snapshot(table({ vote }));
    expect(dialogNamed(dockOf(c), "open vote")).not.toBeNull();
  });
});

describe("game over, in the dock", () => {
  it("Back to lobby is the dock's primary; the banner stays in the strip without it", async () => {
    const c = await mountGame(
      table({
        state: "ended",
        seats: [seat(ME, "Me", 0), { ...seat(OPP, "Opp", 1), eliminated: true } as PlayerView],
      } as Partial<GameView>),
    );
    const dlg = dialogNamed(dockOf(c), "game over")!;
    expect(dlg).not.toBeNull();
    expect(dlg.dataset.rank).toBe("gameOver");
    const back = buttons(dlg, "Back to lobby");
    expect(back).toHaveLength(1);
    expect(back[0].classList.contains("primary")).toBe(true);
    // It takes the bar: no next, no Pass turn.
    expect(buttons(c, "next")).toHaveLength(0);
    // The banner is still in the strip, without its button.
    const strip = stripOf(c);
    expect(strip.querySelector(".game-end")).not.toBeNull();
    expect(buttons(strip, "Back to lobby")).toHaveLength(0);
    // Enter does not leave the table.
    keydown("Enter");
    expect(location.hash).not.toBe("#/lobby");
    click(back[0]);
    expect(location.hash).toBe("#/lobby");
  });
});
