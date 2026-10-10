// #2880: a pending choice's permanents are picked on the board as well
// as in its sheet. These are the pure halves: which choices the board
// can answer, the one shared selection, and the confirm's enable rule.

import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import {
  boardChoicePick,
  boardPickEligible,
  canConfirmBoardPick,
  isBoardPickKind,
  isBoardPickable,
  isBoardPicked,
  pickOnBoard,
  publishBoardPick,
  selectionLegal,
  toggleBoardPick,
  type BoardChoicePick,
} from "./boardChoicePick";
import type { CardView, GameView, PendingChoiceView } from "./protocol";

const perm = (id: string): CardView =>
  ({ instance_id: id, name: id, owner: "me", controller: "me" }) as CardView;

const board = (ids: string[]): Pick<GameView, "battlefield"> =>
  ({ battlefield: { kind: "battlefield", count: ids.length, cards: ids.map(perm) } }) as Pick<
    GameView,
    "battlefield"
  >;

const choice = (over: Partial<PendingChoiceView>): PendingChoiceView =>
  ({
    id: "c1",
    kind: "sacrifice_choice",
    chooser: "me",
    from_player: "me",
    count: 1,
    options: [perm("nazgul-1"), perm("nazgul-2")],
    ...over,
  }) as PendingChoiceView;

const pick = (over: Partial<BoardChoicePick> = {}): BoardChoicePick => ({
  choiceID: "c1",
  eligible: new Set(["a", "b", "c"]),
  selected: new Set(),
  min: 1,
  max: 1,
  ...over,
});

beforeEach(() => boardChoicePick.set(null));

describe("which choices the board answers", () => {
  it("covers the sacrifice and choose-a-permanent kinds", () => {
    for (const k of [
      "sacrifice_choice",
      "entry_sacrifice",
      "their_permanents",
      "own_permanents",
      "ring_bearer",
      "proliferate",
      "copy_target",
      "choose_source",
      "choose_cards",
      "untap_choice",
    ]) {
      expect(isBoardPickKind(k), k).toBe(true);
    }
  });

  it("leaves the hand, library and target kinds alone", () => {
    for (const k of [
      "discard_from_hand",
      "search_library",
      "reveal_pick",
      "pick_target",
      "legend_rule",
      "pay_unless",
      "damage_assignment",
    ]) {
      expect(isBoardPickKind(k), k).toBe(false);
    }
  });

  it("offers the options that are on the battlefield", () => {
    const got = boardPickEligible(choice({}), board(["nazgul-1", "nazgul-2", "forest"]));
    expect(got).toEqual(new Set(["nazgul-1", "nazgul-2"]));
  });

  it("keeps only the options on the battlefield (a source on the stack is not)", () => {
    const c = choice({ kind: "choose_source", options: [perm("bolt"), perm("ogre")] });
    expect(boardPickEligible(c, board(["ogre"]))).toEqual(new Set(["ogre"]));
  });

  it("is null when no option is on the battlefield, or the kind is another", () => {
    expect(boardPickEligible(choice({}), board(["forest"]))).toBeNull();
    expect(
      boardPickEligible(choice({ kind: "discard_from_hand" }), board(["nazgul-1"])),
    ).toBeNull();
    expect(boardPickEligible(null, board(["nazgul-1"]))).toBeNull();
  });
});

describe("a click on the board", () => {
  it("picks an offered permanent, and a second click puts it back", () => {
    const one = toggleBoardPick(pick(), "a");
    expect([...one.selected]).toEqual(["a"]);
    const none = toggleBoardPick(one, "a");
    expect(none.selected.size).toBe(0);
  });

  it("stops at the ceiling, as the sheet's grid does", () => {
    const full = pick({ selected: new Set(["a"]) });
    expect(toggleBoardPick(full, "b")).toBe(full);
    const two = toggleBoardPick(pick({ max: 2 }), "a");
    expect([...toggleBoardPick(two, "b").selected]).toEqual(["a", "b"]);
  });

  it("does nothing on a permanent the choice does not offer", () => {
    const s = pick();
    expect(toggleBoardPick(s, "zz")).toBe(s);
  });

  it("goes through the shared store, and is swallowed while a choice is open", () => {
    expect(pickOnBoard("a")).toBe(false);
    boardChoicePick.set(pick());
    expect(pickOnBoard("a")).toBe(true);
    expect(isBoardPicked(get(boardChoicePick), "a")).toBe(true);
    // A permanent it does not offer: handled (not clickable), unchanged.
    expect(pickOnBoard("zz")).toBe(true);
    expect([...get(boardChoicePick)!.selected]).toEqual(["a"]);
  });

  it("highlights only what the choice offers", () => {
    const s = pick();
    expect(isBoardPickable(s, "a")).toBe(true);
    expect(isBoardPickable(s, "zz")).toBe(false);
    expect(isBoardPickable(null, "a")).toBe(false);
  });
});

describe("the dock's confirm", () => {
  it("enables between the floor and the ceiling", () => {
    expect(selectionLegal(0, 1, 1)).toBe(false);
    expect(selectionLegal(1, 1, 1)).toBe(true);
    expect(selectionLegal(0, 0, 3)).toBe(true);
    expect(selectionLegal(4, 0, 3)).toBe(false);
  });

  it("follows the board's picks", () => {
    expect(canConfirmBoardPick(null)).toBe(false);
    const s = pick({ min: 2, max: 2 });
    expect(canConfirmBoardPick(s)).toBe(false);
    const one = toggleBoardPick(s, "a");
    expect(canConfirmBoardPick(one)).toBe(false);
    expect(canConfirmBoardPick(toggleBoardPick(one, "b"))).toBe(true);
  });
});

describe("publishing", () => {
  it("writes only a change, so the sheet and the board can follow each other", () => {
    const s = pick({ selected: new Set(["a"]) });
    publishBoardPick(s);
    const first = get(boardChoicePick);
    publishBoardPick(pick({ selected: new Set(["a"]) }));
    expect(get(boardChoicePick)).toBe(first);
    publishBoardPick(pick({ selected: new Set(["b"]) }));
    expect([...get(boardChoicePick)!.selected]).toEqual(["b"]);
    publishBoardPick(null);
    expect(get(boardChoicePick)).toBeNull();
  });
});
