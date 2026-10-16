// @vitest-environment jsdom
//
// reflow.render.test.ts — ADR 0077 §6 as amended 2026-10-16 (#4065):
// the "reflow" expand style on a real Board with four seats in the row
// layout. On an opponent's turn their board takes most of the row and
// the other two are avatar cards; on your own turn, with the overlay
// style, or with the active-player expansion off, the row is three
// equal boards; a seat a live prompt needs is never an avatar card; and
// an avatar card's avatar still opens the ADR 0120 overlay. The decision
// itself is pinned in expansion.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import { begin, targeting } from "./targeting";
import { _resetForTests as resetDock } from "./dock";
import { resetSettings, updateSettings } from "./settings";
import { findStandIn } from "./boardAnchor";
import type { CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { cleanup, flushSync, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  vi.useFakeTimers();
  resetSettings();
  updateSettings("display", "expandStyle", "reflow");
});

afterEach(() => {
  cleanup();
  targeting.set(null);
  resetDock();
  resetSettings();
  vi.useRealTimers();
});

const ME = "me";
const BOB = "bob";
const CAT = "cat";
const DEE = "dee";

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

const bear = (id: string, controller: string): CardView =>
  ({
    instance_id: id,
    name: "Grizzly Bears",
    owner: controller,
    controller,
    type_line: "Creature — Bear",
    power: 2,
    toughness: 2,
  }) as unknown as CardView;

// A spell that can target Cat's bear and nothing else.
const pinpoint = (): CardView =>
  ({
    instance_id: "pinpoint",
    name: "Pinpoint",
    owner: ME,
    controller: ME,
    type_line: "Instant",
    legal_targets: { min: 1, max: 1, players: [], cards: ["cat-bear"] },
  }) as unknown as CardView;

// Seats in turn order: Bob sits next (left), Cat across, Dee across_next.
const gameView = (activeSeat: number): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(BOB, "Bob", 1), seat(CAT, "Cat", 2), seat(DEE, "Dee", 3)],
    battlefield: zone("battlefield", undefined, [
      bear("bob-bear", BOB),
      bear("cat-bear", CAT),
      bear("dee-bear", DEE),
    ]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      number: 3,
      active_seat: activeSeat,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
  }) as unknown as GameView;

function mountBoard(activeSeat: number, view: GameView = gameView(activeSeat)) {
  return render(
    Board as never,
    {
      view,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
}

const board = () => document.querySelector<HTMLElement>(".board")!;
const region = (name: string) =>
  document.querySelector<HTMLElement>(`.board > .slot [role="region"][aria-label="${name} board"]`);
const isAvatarCard = (name: string) => region(name)?.classList.contains("avatar-card") ?? false;
const isFullBoard = (name: string) => region(name)?.classList.contains("panel") ?? false;
const avatarCards = () => [...document.querySelectorAll(".board > .slot .avatar-card")];

describe("reflow on an opponent's turn", () => {
  it("gives Bob most of the row and shrinks Cat and Dee to avatar cards", () => {
    mountBoard(1);
    expect(isFullBoard("Bob")).toBe(true);
    expect(isAvatarCard("Cat")).toBe(true);
    expect(isAvatarCard("Dee")).toBe(true);
    // The row's columns: Bob's two shares, then two fixed avatar columns.
    const cols = board().style.gridTemplateColumns;
    expect(cols).toContain("minmax(0, 2fr)");
    expect(cols.match(/--reflow-avatar-w/g)).toHaveLength(2);
    expect(board().hasAttribute("data-reflow")).toBe(true);
    // Your own board is untouched.
    expect(document.querySelector(".board > .slot[data-pos='self'] .avatar-card")).toBeNull();
  });

  it("follows the turn to Dee, at the right end of the row", () => {
    mountBoard(3);
    expect(isFullBoard("Dee")).toBe(true);
    expect(isAvatarCard("Bob")).toBe(true);
    expect(isAvatarCard("Cat")).toBe(true);
    expect(board().style.gridTemplateColumns).toMatch(/2fr\)$/);
  });

  it("keeps each avatar card the seat's board region, named, with its life and avatar", () => {
    mountBoard(1);
    const cat = region("Cat")!;
    expect(cat.getAttribute("role")).toBe("region");
    expect(cat.querySelector(`[data-seat-id="${CAT}"]`)).not.toBeNull();
    expect(cat.textContent).toContain("Cat");
    expect(cat.textContent).toContain("40");
    // The keyboard route to the overlay is still there.
    expect(cat.querySelector('button[aria-label="Show Cat\'s full board"]')).not.toBeNull();
    // It is the stand-in an arrow to one of Cat's cards ends on.
    expect(findStandIn(document, CAT)).toBe(cat);
    expect(findStandIn(document, BOB)).toBeNull();
  });

  it("applies under summaries too: the active opponent is a full board", () => {
    updateSettings("display", "opponentDetail", "summary");
    mountBoard(1);
    expect(isFullBoard("Bob")).toBe(true);
    expect(avatarCards()).toHaveLength(2);
  });

  it("hovering an avatar card's avatar opens that seat's board in the overlay", () => {
    mountBoard(1);
    const av = region("Cat")!.querySelector<HTMLElement>(`[data-seat-id="${CAT}"]`)!;
    av.dispatchEvent(new MouseEvent("pointerenter"));
    flushSync();
    vi.advanceTimersByTime(400);
    flushSync();
    const o = document.querySelector<HTMLElement>("[data-board-expanded]")!;
    expect(o.getAttribute("aria-label")).toBe("Cat's board, expanded");
    expect(o.querySelector('[data-instance-id="cat-bear"]')).not.toBeNull();
  });

  it("never shrinks a seat a live prompt needs", () => {
    mountBoard(1);
    begin(pinpoint(), "any");
    flushSync();
    expect(isFullBoard("Cat")).toBe(true);
    expect(region("Cat")!.parentElement!.querySelector('[data-instance-id="cat-bear"]')).not.toBe(
      null,
    );
    expect(isAvatarCard("Dee")).toBe(true);
    targeting.set(null);
    flushSync();
    expect(isAvatarCard("Cat")).toBe(true);
  });
});

describe("no reflow", () => {
  function expectEqualRow(): void {
    expect(avatarCards()).toHaveLength(0);
    expect(board().style.gridTemplateColumns).toBe("");
    expect(board().hasAttribute("data-reflow")).toBe(false);
    for (const name of ["Bob", "Cat", "Dee"]) expect(isFullBoard(name)).toBe(true);
  }

  it("on your own turn: three equal boards", () => {
    mountBoard(0);
    expectEqualRow();
  });

  it("with the overlay style, which keeps the boards equal", () => {
    updateSettings("display", "expandStyle", "overlay");
    mountBoard(1);
    expectEqualRow();
  });

  it("with the active-player expansion off", () => {
    updateSettings("display", "expandActivePlayer", false);
    mountBoard(1);
    expectEqualRow();
  });

  it("at a four-player quadrant table, which has no one opponents' row", () => {
    updateSettings("display", "tableLayout", "quadrant");
    mountBoard(1);
    expectEqualRow();
  });
});

// The stack pile docks on a seam of the top row (#3380). With the
// reflow's avatar cards in the row it stays over the full board instead
// (stackPile.ts, tableSeams().dock), and measuring that span must not
// feed back into the effect that measures it.
describe("the stack pile beside the avatar cards", () => {
  const BOARD = { w: 1400, h: 800 };
  const AV = 132;
  let restore: (() => void) | null = null;
  afterEach(() => {
    restore?.();
    restore = null;
  });

  // jsdom lays nothing out, so the board and its slots report the boxes
  // the row layout gives them on Dee's turn: two avatar cards, then Dee.
  function layOut(): void {
    const proto = HTMLElement.prototype;
    const original = proto.getBoundingClientRect;
    proto.getBoundingClientRect = function (this: HTMLElement): DOMRect {
      const box = (left: number, top: number, width: number, height: number) =>
        ({
          left,
          top,
          width,
          height,
          right: left + width,
          bottom: top + height,
          x: left,
          y: top,
        }) as DOMRect;
      if (this.classList.contains("board")) return box(0, 0, BOARD.w, BOARD.h);
      if (this.classList.contains("slot")) {
        const half = BOARD.h / 2;
        switch (this.dataset.pos) {
          case "next":
            return box(0, 0, AV, half);
          case "across":
            return box(AV + 8, 0, AV, half);
          case "across_next":
            return box(2 * (AV + 8), 0, BOARD.w - 2 * (AV + 8), half);
          case "self":
            return box(0, half + 8, BOARD.w, half - 8);
        }
      }
      return original.call(this);
    };
    restore = () => {
      proto.getBoundingClientRect = original;
    };
  }

  it("mounts with a spell on the stack and docks clear of the avatar cards", () => {
    updateSettings("display", "stackStyle", "pile");
    layOut();
    const item = {
      id: "s1",
      kind: "spell",
      controller: DEE,
      owner: DEE,
      source_card_id: "s1",
      targets: [],
    } as unknown as StackItemView;
    const spell = {
      instance_id: "s1",
      name: "Lightning Bolt",
      owner: DEE,
      controller: DEE,
      type_line: "Instant",
    } as unknown as CardView;
    const view = { ...gameView(3), stack: zone("stack", undefined, [spell]), stack_items: [item] };
    expect(() => {
      mountBoard(3, view as GameView);
      flushSync();
    }).not.toThrow();
    expect(isAvatarCard("Bob")).toBe(true);
    expect(isAvatarCard("Cat")).toBe(true);
    const lane = document.querySelector<HTMLElement>(".stack-lane");
    expect(lane).not.toBeNull();
  });
});
