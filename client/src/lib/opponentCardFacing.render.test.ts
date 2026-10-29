// @vitest-environment jsdom
//
// opponentCardFacing.render.test.ts — #4555: which way the cards on a
// top-row (flipped) opponent's panel face.
//
// jsdom does no layout and applies no component CSS, so these pin the
// contract the CSS draws by (the panel's class, the pile index it
// stacks by) and read the stylesheet's source for the two halves of the
// rule: the rows stay turned 180° and every pile in them turns back.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import panelSource from "./components/board/PlayerPanel.svelte?raw";
import { defaultSettings, resetSettings, settings, updateSettings } from "./settings";
import {
  DEFAULT_OPPONENT_CARD_FACING,
  isOpponentCardFacing,
  OPPONENT_CARD_FACINGS,
} from "./opponentCardFacing";
import type { ActionType, CardView, GameView, PlayerView } from "./protocol";
import { cleanup, flushSync, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  localStorage.clear();
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

const creature = (id: string, owner: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: id,
    owner,
    controller: owner,
    known_by_you: true,
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
    hand: zone("hand", id, [], id === ME ? 0 : 3),
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
    seats: [seat(ME, "Me", 0), seat(BOB, "Bob", 1), seat(CAT, "Cat", 2), seat(DEE, "Dee", 3)],
    battlefield: zone("battlefield", undefined, [
      creature("bob-bear", BOB),
      creature("bob-ogre", BOB, { tapped: true, name: "Ogre", power: 4, toughness: 3 }),
      creature("bob-wolf", BOB, { name: "Wolf", power: 3, toughness: 3 }),
      creature("me-bear", ME),
    ]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 1, priority_holder: 0, phase: "combat", step: "attackers" },
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

describe("the setting", () => {
  it("defaults to them, today's look", () => {
    expect(DEFAULT_OPPONENT_CARD_FACING).toBe("them");
    expect(defaultSettings().display.opponentCardFacing).toBe("them");
    expect(get(settings).display.opponentCardFacing).toBe("them");
  });

  it("knows exactly its two values", () => {
    expect(OPPONENT_CARD_FACINGS).toEqual(["them", "me"]);
    expect(isOpponentCardFacing("me")).toBe(true);
    expect(isOpponentCardFacing("them")).toBe(true);
    expect(isOpponentCardFacing("sideways")).toBe(false);
    expect(isOpponentCardFacing(undefined)).toBe(false);
  });
});

describe("a flipped panel's cards", () => {
  it("face their owner by default: no upright class on any panel", () => {
    updateSettings("display", "tableLayout", "row");
    mountBoard();
    for (const id of [ME, BOB, CAT, DEE]) {
      expect(panelOf(id)!.classList.contains("faces-viewer"), id).toBe(false);
    }
    // The opponents are still the flipped panels the turned rows belong to.
    expect(panelOf(BOB)!.classList.contains("flipped")).toBe(true);
  });

  it("face the viewer on every flipped opponent's panel when asked, and on no other", () => {
    updateSettings("display", "tableLayout", "row");
    updateSettings("display", "opponentCardFacing", "me");
    mountBoard();
    for (const id of [BOB, CAT, DEE]) {
      expect(panelOf(id)!.classList.contains("flipped"), id).toBe(true);
      expect(panelOf(id)!.classList.contains("faces-viewer"), id).toBe(true);
    }
    expect(panelOf(ME)!.classList.contains("faces-viewer")).toBe(false);
  });

  it("leave an upright panel alone: the quadrant's next seat is not flipped", () => {
    updateSettings("display", "tableLayout", "quadrant");
    updateSettings("display", "opponentCardFacing", "me");
    mountBoard();
    const next = document.querySelector<HTMLElement>('.board > .slot[data-pos="next"] .panel')!;
    expect(next.classList.contains("flipped")).toBe(false);
    expect(next.classList.contains("faces-viewer")).toBe(false);
    // The across-table seats are flipped in the quadrant too.
    expect(panelOf(CAT)!.classList.contains("faces-viewer")).toBe(true);
  });

  it("follow the setting live, and the board keeps its rows either way", () => {
    updateSettings("display", "tableLayout", "row");
    mountBoard();
    const rows = () =>
      [...panelOf(BOB)!.querySelectorAll(".row-cards .pile")].map((p) =>
        p.querySelector("[data-instance-id]")?.getAttribute("data-instance-id"),
      );
    const before = rows();
    expect(before.length).toBeGreaterThan(0);
    updateSettings("display", "opponentCardFacing", "me");
    flushSync();
    expect(panelOf(BOB)!.classList.contains("faces-viewer")).toBe(true);
    expect(rows()).toEqual(before);
    updateSettings("display", "opponentCardFacing", "them");
    flushSync();
    expect(panelOf(BOB)!.classList.contains("faces-viewer")).toBe(false);
  });

  it("number their piles in row order, so the turned-back piles stack by it", () => {
    updateSettings("display", "tableLayout", "row");
    updateSettings("display", "opponentCardFacing", "me");
    mountBoard();
    const piles = [...panelOf(BOB)!.querySelectorAll<HTMLElement>(".row-cards .pile")];
    const creatures = piles.filter((p) => p.closest(".grid-creatures"));
    expect(creatures.length).toBeGreaterThan(1);
    expect(creatures.map((p) => p.style.getPropertyValue("--pi"))).toEqual(
      creatures.map((_, i) => String(i)),
    );
  });
});

describe("the stylesheet", () => {
  const src = panelSource;
  const faces = ".panel.opponent.flipped.faces-viewer";

  it("keeps the rows turned 180° and turns each pile back, so only the cards move", () => {
    expect(src).toContain(
      ".panel.opponent.flipped :is(.grid-creatures, .grid-middle) :global(.row-cards) {\n      transform: rotate(180deg);",
    );
    expect(src).toMatch(
      new RegExp(
        String.raw`${faces.replaceAll(".", "\\.")} :is\(\.grid-creatures, \.grid-middle\) :global\(\.pile\) \{\s+transform: rotate\(180deg\);`,
      ),
    );
  });

  it("draws the pile tops and the commander upright, and leaves the hand turned", () => {
    expect(src).toMatch(
      new RegExp(
        String.raw`${faces.replaceAll(".", "\\.")} \.rail :global\(\.pile \.thumb\),\s+${faces.replaceAll(".", "\\.")} :global\(\.command-strip \.slot \.card\) \{\s+transform: none;`,
      ),
    );
    expect(src).toContain(
      ".panel.opponent.flipped .hand-zone > :global(.hand),\n    .panel.opponent.flipped .rail :global(.pile .thumb),\n    .panel.opponent.flipped :global(.command-strip .slot .card) {\n      transform: rotate(180deg);",
    );
    expect(src).not.toMatch(/faces-viewer[^{]*\.hand-zone/);
  });
});
