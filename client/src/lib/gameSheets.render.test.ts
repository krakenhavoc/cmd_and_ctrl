// @vitest-environment jsdom
//
// gameSheets.render.test.ts — ADR 0111 Delivery PR 6 (S56, #1958),
// through the real Game route. The two sheets the game itself waits
// on: the opening hand (owner decision 2: the hand in a sheet that
// grows up out of the dock, Keep hand / Mulligan in its action bar) and
// the discard to hand size (DiscardPromptModal). For each: it opens as a
// sheet inside region "actions", in one non-modal dialog that keeps the
// name the e2e suite finds it by; its buttons are the dock's bar and
// send the right action; Enter and Escape do what ADR 0111 §1 says; it
// minimises and restores; and nothing of it is a centred modal. The
// board is a stub that renders the strip Game.svelte hands it.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import { PROTOCOL_VERSION, type CardView, type GameView, type PlayerView } from "./protocol";
import { session, type Session } from "./session";
import { _resetForTests as resetDock, SHEET_HAND_WIDTH } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { targeting, setConfirmHandler } from "./targeting";
import { defaultSettings, settings } from "./settings";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import { nameOf } from "./test/dockView";

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
  payload?: { type?: string; player?: string; params?: Record<string, unknown> };
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
  actions(type: string): SentFrame[] {
    return this.sent.filter((f) => f.kind === "action" && f.payload?.type === type);
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

const card = (id: string, name: string): CardView =>
  ({ instance_id: id, name, owner: ME, controller: ME }) as CardView;

const HAND = [
  card("h1", "Forest"),
  card("h2", "Island"),
  card("h3", "Sol Ring"),
  card("h4", "Lightning Bolt"),
  card("h5", "Mulldrifter"),
  card("h6", "Swamp"),
  card("h7", "Plains"),
];

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

function table(over: Partial<GameView> = {}, me: Partial<PlayerView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, "Me", 0, me), seat(OPP, "Opp", 1)],
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
    ...over,
  } as unknown as GameView;
}

// The mulligan window, the viewer's seven cards undecided.
const mulliganTable = (taken = 0) =>
  table({ mulligans_open: true, turn: { ...table().turn, number: 1, step: "untap" } }, {
    hand: zone("hand", ME, HAND),
    hand_kept: false,
    mulligans_taken: taken,
  } as never);

// Cleanup with a hand of eight and one card owed.
const discardTable = () =>
  table(
    {
      discard_pending: { [ME]: 1 },
      turn: { ...table().turn, step: "cleanup", phase: "ending", priority_holder: -1 },
    } as never,
    { hand: zone("hand", ME, [...HAND, card("h8", "Opt")]), max_hand_size: 7 } as never,
  );

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
const dialogNamed = (root: ParentNode, name: string): HTMLElement | null =>
  [...root.querySelectorAll<HTMLElement>('[role="dialog"]')].find(
    (d) => d.getAttribute("aria-label") === name,
  ) ?? null;
const buttonNamed = (root: ParentNode, name: string): HTMLButtonElement | null =>
  [...root.querySelectorAll<HTMLButtonElement>("button")].find((b) => nameOf(b) === name) ?? null;
function keydown(key: string, target: EventTarget = document.body): void {
  target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true }));
  flushSync();
}
function blur(): void {
  (document.activeElement as HTMLElement | null)?.blur?.();
}

// The sheet is in the dock, in one non-modal dialog of that name, and
// there is no backdrop or centred modal anywhere on the page.
function expectSheet(c: HTMLElement, name: string): HTMLElement {
  const dlg = dialogNamed(dockOf(c), name);
  expect(dlg, `dialog "${name}" in the dock`).not.toBeNull();
  expect(dlg!.getAttribute("aria-modal")).toBeNull();
  expect(dlg!.dataset.rank).toBe("choice");
  const sheet = dlg!.querySelector<HTMLElement>(".dock-sheet");
  expect(sheet, "a sheet panel inside the dialog").not.toBeNull();
  expect(sheet!.hidden).toBe(false);
  expect(c.querySelector(".prompt-backdrop")).toBeNull();
  expect(c.querySelector(".prompt-modal")).toBeNull();
  expect(c.querySelector('[aria-modal="true"]')).toBeNull();
  return dlg!;
}

describe("the opening hand, as a dock sheet (decision 2)", () => {
  it("opens as a sheet named 'keep or mulligan your hand', with the seven cards in it", async () => {
    const c = await mountGame(mulliganTable());
    const dlg = expectSheet(c, "keep or mulligan your hand");
    const list = dlg.querySelector('[role="list"][aria-label="your opening hand"]')!;
    expect(list).not.toBeNull();
    expect(list.querySelectorAll('[role="listitem"]')).toHaveLength(7);
    expect(dlg.textContent).toContain("Hand size: 7");
    expect(dlg.querySelector(".sheet-title")?.textContent?.trim()).toBe("Your opening hand");
    // The old centred dialog and its scrim are gone.
    expect(c.querySelector(".mulligan-dialog")).toBeNull();
    expect(c.querySelector(".mulligan-scrim")).toBeNull();
    // The roll call stays in the strip, outside the dock.
    expect(dockOf(c).querySelector('[aria-label="opening hand decisions"]')).toBeNull();
  });

  it("asks for a wide sheet, so the seven cards are full size (#2200)", async () => {
    const c = await mountGame(mulliganTable());
    const dlg = expectSheet(c, "keep or mulligan your hand");
    const sheet = dlg.querySelector<HTMLElement>(".dock-sheet")!;
    // jsdom has no layout: the contract is the width the sheet asks for
    // (CSS caps it at the screen less 24px). 720px gave ~95px cards.
    expect(sheet.style.getPropertyValue("--sheet-want")).toBe(`${SHEET_HAND_WIDTH}px`);
    expect(SHEET_HAND_WIDTH).toBeGreaterThan(7 * 200);
    // The hand is still in the sheet, and the buttons are still in the bar.
    expect(sheet.querySelectorAll(".mulligan-card")).toHaveLength(7);
    expect(buttonNamed(dlg, "Keep hand")).not.toBeNull();
    expect(buttonNamed(dlg, "Mulligan")).not.toBeNull();
    expect(sheet.contains(buttonNamed(dlg, "Keep hand"))).toBe(false);
  });

  it("puts Keep hand in the corner and Mulligan on the left of the bar, and sends each", async () => {
    const c = await mountGame(mulliganTable());
    const dlg = expectSheet(c, "keep or mulligan your hand");
    const primary = dlg.querySelector<HTMLButtonElement>(".dock-bar .request-primary")!;
    expect(nameOf(primary)).toBe("Keep hand");
    const secondaries = [...dlg.querySelectorAll<HTMLButtonElement>(".dock-bar .secondary")];
    expect(secondaries.map(nameOf)).toEqual(["Mulligan"]);
    // One of each on the page (the e2e suite's locators are strict).
    expect([...c.querySelectorAll("button")].filter((b) => nameOf(b) === "Keep hand")).toHaveLength(
      1,
    );
    // next and Pass turn give way while it is open.
    expect(dockOf(c).querySelector(".dock-btn.next")).toBeNull();

    click(secondaries[0]);
    const mull = FakeSocket.last!.actions("mulligan");
    expect(mull).toHaveLength(1);
    expect(mull[0].payload?.params).toEqual({ hand_size: 7 });
    click(primary);
    expect(FakeSocket.last!.actions("keep_hand")).toHaveLength(1);
  });

  it("keeps on Enter (a confirm of the hand on screen), and Escape does not mulligan", async () => {
    const c = await mountGame(mulliganTable());
    expectSheet(c, "keep or mulligan your hand");
    blur();
    keydown("Escape");
    expect(FakeSocket.last!.actions("mulligan")).toHaveLength(0);
    expect(FakeSocket.last!.actions("keep_hand")).toHaveLength(0);
    keydown("Enter");
    expect(FakeSocket.last!.actions("keep_hand")).toHaveLength(1);
  });

  it("takes focus into the sheet when it opens", async () => {
    const c = await mountGame(mulliganTable());
    const dlg = expectSheet(c, "keep or mulligan your hand");
    expect(document.activeElement).toBe(dlg.querySelector(".dock-sheet"));
  });

  it("minimises to a restore chip with Keep hand still in the bar, and restores", async () => {
    const c = await mountGame(mulliganTable());
    const dlg = expectSheet(c, "keep or mulligan your hand");
    click(dlg.querySelector<HTMLButtonElement>("button.sheet-min")!);
    expect(dlg.querySelector<HTMLElement>(".dock-sheet")!.hidden).toBe(true);
    const chip = buttonNamed(dlg, "restore: Your opening hand")!;
    expect(chip).not.toBeNull();
    expect(buttonNamed(dlg, "Keep hand")).not.toBeNull();
    click(chip);
    expect(dlg.querySelector<HTMLElement>(".dock-sheet")!.hidden).toBe(false);
  });

  it("closes once the hand is kept, and next comes back", async () => {
    const c = await mountGame(mulliganTable());
    expectSheet(c, "keep or mulligan your hand");
    snapshot(table());
    expect(dialogNamed(c, "keep or mulligan your hand")).toBeNull();
    expect(dockOf(c).querySelector(".dock-btn.next")).not.toBeNull();
  });
});

// The e2e scry caught this: in Game.svelte the dock comes BEFORE
// ChoicePromptModal, so a sheet body the dock rendered as a snippet
// updated before the modal's `{#if}` closed it, and read the prompt
// after it was gone (`active.count` of null). The body is the picker's
// own DOM now, moved into the dock, so it closes with its `{#if}`.
describe("a pending-choice sheet closing on a new frame (through the Game route)", () => {
  const scryTable = () =>
    table({
      pending_choices: [
        {
          id: "scry-1",
          kind: "scry",
          chooser: ME,
          from_player: ME,
          count: 1,
          reason: "Scry 1",
          options: [card("top", "Forest")],
        },
      ],
    } as never);

  it("answers the scry from the bar, then closes cleanly when the prompt goes", async () => {
    const errors: unknown[] = [];
    const onError = (e: ErrorEvent) => errors.push(e.error ?? e.message);
    window.addEventListener("error", onError);
    try {
      const c = await mountGame(scryTable());
      const dlg = expectSheet(c, "Scry 1");
      click(buttonNamed(dlg, "put Forest on the bottom")!);
      click(dlg.querySelector<HTMLButtonElement>(".dock-bar .request-primary")!);
      const sent = FakeSocket.last!.actions("resolve_choice");
      expect(sent).toHaveLength(1);
      expect(sent[0].payload?.params).toEqual({
        choice_id: "scry-1",
        bottom: ["top"],
        top_order: [],
      });
      // The server drains the prompt.
      expect(() => snapshot(table())).not.toThrow();
      expect(dialogNamed(c, "Scry 1")).toBeNull();
      expect(c.querySelector(".dock-sheet")).toBeNull();
      expect(dockOf(c).querySelector(".dock-btn.next")).not.toBeNull();
      expect(errors).toEqual([]);
    } finally {
      window.removeEventListener("error", onError);
    }
  });
});

describe("the discard to hand size, as a dock sheet", () => {
  it("opens as a sheet named 'Discard 1 card' with the hand as its grid", async () => {
    const c = await mountGame(discardTable());
    const dlg = expectSheet(c, "Discard 1 card");
    expect(dlg.querySelectorAll(".dock-sheet button.card-pick")).toHaveLength(8);
    // The e2e suite reads the owed count off `.prompt-count`.
    expect(dlg.querySelector(".prompt-count")?.textContent?.trim()).toBe("0 / 1 selected");
  });

  it("holds Discard until the picks add up, then sends discard_selection", async () => {
    const c = await mountGame(discardTable());
    const dlg = expectSheet(c, "Discard 1 card");
    const discard = dlg.querySelector<HTMLButtonElement>(".dock-bar .request-primary")!;
    expect(nameOf(discard)).toBe("Discard");
    expect(discard.disabled).toBe(true);
    // No cancel: the discard is owed.
    expect(dlg.querySelectorAll(".dock-bar .secondary")).toHaveLength(0);

    const pick = dlg.querySelector<HTMLButtonElement>('button.card-pick[aria-label="select Opt"]')!;
    click(pick);
    expect(dlg.querySelector(".prompt-count")?.textContent?.trim()).toBe("1 / 1 selected");
    // The sheet was re-drawn in place, not re-mounted: the same button.
    expect(dlg.querySelector('button.card-pick[aria-label="select Opt"]')).toBe(pick);
    expect(discard.disabled).toBe(false);
    click(discard);
    const sent = FakeSocket.last!.actions("discard_selection");
    expect(sent).toHaveLength(1);
    expect(sent[0].payload?.params).toEqual({ card_ids: ["h8"] });
  });

  it("presses Discard on Enter only once it is ready; Escape does nothing", async () => {
    const c = await mountGame(discardTable());
    const dlg = expectSheet(c, "Discard 1 card");
    blur();
    keydown("Enter");
    keydown("Escape");
    expect(FakeSocket.last!.actions("discard_selection")).toHaveLength(0);
    click(dlg.querySelector<HTMLButtonElement>('button.card-pick[aria-label="select Opt"]')!);
    blur();
    keydown("Enter");
    expect(FakeSocket.last!.actions("discard_selection")).toHaveLength(1);
  });
});
