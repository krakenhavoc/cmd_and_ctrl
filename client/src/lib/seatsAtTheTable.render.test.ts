// @vitest-environment jsdom
//
// seatsAtTheTable.render.test.ts — ADR 0204: every seat's identity sits
// at the middle of its hand's edge, in front of the fan.
//
// jsdom does no layout, so these pin the contract the CSS places by: the
// identity is the panel's seated one (class and attribute, never pixels),
// the hand count rides on an opponent's portrait, and the viewer's fan
// takes the notch it dips under.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import Hand from "./components/board/Hand.svelte";
import { resetSettings, updateSettings } from "./settings";
import type { ActionType, CardView, GameView, PlayerView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
});

afterEach(() => {
  cleanup();
  resetSettings();
});

const ME = "me";
const BOB = "bob";
const CAT = "cat";
const DEE = "dee";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = [], count?: number) => ({
  kind,
  owner,
  count: count ?? cards.length,
  cards,
});

const card = (id: string, owner: string): CardView =>
  ({ instance_id: id, name: "Plains", owner, controller: owner, known_by_you: true }) as CardView;

const seat = (id: string, name: string, n: number, inHand: number): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand:
      id === ME
        ? zone(
            "hand",
            id,
            Array.from({ length: inHand }, (_, i) => card(`h${i}`, id)),
          )
        : zone("hand", id, [], inHand),
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
    seats: [
      seat(ME, "Me", 0, 3),
      seat(BOB, "Bob", 1, 7),
      seat(CAT, "Cat", 2, 1),
      seat(DEE, "Dee", 3, 0),
    ],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "precombat_main" },
    mulligans_open: false,
  }) as unknown as GameView;

function mountBoard() {
  return render(
    Board as never,
    {
      view: gameView(),
      viewerID: ME,
      isAdmin: false,
      sendAction: (_type: ActionType) => {},
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
}

const panelOf = (seatID: string) =>
  document
    .querySelector(`.board > .slot [data-seat-identity="${seatID}"]`)
    ?.closest<HTMLElement>(".panel");
const identityOf = (seatID: string) =>
  document.querySelector<HTMLElement>(`.board > .slot [data-seat-identity="${seatID}"]`);

describe("every seat is seated at its hand's edge (§1)", () => {
  it("each table panel's identity is the seated one, in its own panel's rail", () => {
    mountBoard();
    for (const id of [ME, BOB, CAT, DEE]) {
      const identity = identityOf(id);
      expect(identity, id).not.toBeNull();
      expect(identity!.classList.contains("seated"), id).toBe(true);
      expect(identity!.parentElement!.classList.contains("rail"), id).toBe(true);
      // The disc keeps its data-seat-id: arrows, the trace and the
      // expand peek follow it.
      expect(identity!.querySelector(`[data-seat-id="${id}"]`), id).not.toBeNull();
    }
  });

  it("opponents in the row layout sit at the top edge (flipped), you at the bottom", () => {
    updateSettings("display", "tableLayout", "row");
    mountBoard();
    for (const id of [BOB, CAT, DEE]) {
      expect(panelOf(id)!.classList.contains("flipped"), id).toBe(true);
    }
    expect(panelOf(ME)!.classList.contains("flipped")).toBe(false);
    expect(panelOf(ME)!.classList.contains("self")).toBe(true);
  });

  it("an upright opponent (the quadrant's next seat) is seated at its bottom edge", () => {
    updateSettings("display", "tableLayout", "quadrant");
    mountBoard();
    const next = document.querySelector('.board > .slot[data-pos="next"] .panel')!;
    expect(next.classList.contains("flipped")).toBe(false);
    expect(next.querySelector(".rail > .identity.seated")).not.toBeNull();
  });

  it("the identity keeps its accessible name", () => {
    mountBoard();
    const disc = identityOf(BOB)!.querySelector(`[data-seat-id="${BOB}"]`)!;
    expect(disc.getAttribute("aria-label")).toBe("Bob, 40 life");
  });
});

describe("the hand count beside an opponent's avatar (§2, call 2)", () => {
  it("rides on each opponent's portrait with their count, and not on yours", () => {
    mountBoard();
    const count = (id: string) => identityOf(id)!.querySelector<HTMLElement>(".hand-count");
    expect(count(BOB)!.dataset.handCount).toBe("7");
    expect(count(BOB)!.getAttribute("aria-label")).toBe("7 cards in hand");
    expect(count(CAT)!.getAttribute("aria-label")).toBe("1 card in hand");
    expect(count(DEE)!.dataset.handCount).toBe("0");
    expect(count(ME)).toBeNull();
    // Beside the portrait, in its row, not inside the clickable disc.
    expect(count(BOB)!.closest("[data-seat-id]")).toBeNull();
    expect(count(BOB)!.parentElement!.classList.contains("core-row")).toBe(true);
  });

  it("follows the opponent-hand-count setting", () => {
    updateSettings("display", "showOpponentHandCount", false);
    mountBoard();
    expect(document.querySelectorAll(".board .hand-count")).toHaveLength(0);
  });
});

describe("the viewer's fan dips under the avatar (§3, call 4)", () => {
  const hand = zone(
    "hand",
    ME,
    Array.from({ length: 4 }, (_, i) => card(`c${i}`, ME)),
  );

  it("a card under the notch rests lower by its depth", () => {
    // jsdom lays every slot out at x 0, so a notch over 0 covers all.
    render(
      Hand as never,
      {
        hand,
        isSelf: true,
        notch: { left: -10, right: 10, depth: 18 },
      } as never,
    );
    const slots = [...document.querySelectorAll<HTMLElement>(".hand-slot")];
    expect(slots).toHaveLength(4);
    for (const s of slots) {
      expect(s.style.getPropertyValue("--dip")).toBe("18px");
      expect(s.style.transform).toContain("var(--dip-now, 0px)");
    }
  });

  it("no card dips without a notch, or away from it", () => {
    render(
      Hand as never,
      {
        hand,
        isSelf: true,
        notch: { left: 500, right: 600, depth: 18 },
      } as never,
    );
    for (const s of document.querySelectorAll<HTMLElement>(".hand-slot")) {
      expect(s.style.getPropertyValue("--dip")).toBe("");
    }
  });
});
