// #2394: a card-set pick over permanents on the battlefield ("untap up
// to five lands") is answered by clicking the permanents on the board.
// These are the pure halves: which choices go to the board, the
// targeting state they open, the answer they send, and the dock
// request that asks them.

import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import { answeredOnBoard, listFallback, showChoiceAsList } from "./boardAnsweredChoice";
import { L } from "./labels";
import type { CardView, GameView, PendingChoiceView } from "./protocol";
import {
  beginChoice,
  canConfirm,
  cardSetPickState,
  choiceAnswer,
  isCardSetPick,
  isMultiPick,
  targeting,
  togglePick,
} from "./targeting";
import { targetingRequest } from "./targetingDock";

const land = (id: string): CardView =>
  ({
    instance_id: id,
    name: "Forest",
    owner: "me",
    controller: "me",
    tapped: true,
  }) as CardView;

const finale = {
  instance_id: "finale",
  name: "Finale of Revelation",
  owner: "me",
  controller: "me",
} as CardView;

const untapLands = (over: Partial<PendingChoiceView> = {}): PendingChoiceView =>
  ({
    id: "untap-1",
    kind: "choose_cards",
    chooser: "me",
    from_player: "me",
    source: "finale",
    reason: "Finale of Revelation — untap up to five lands",
    choose_min: 0,
    choose_max: 5,
    options: [land("f1"), land("f2"), land("f3")],
    ...over,
  }) as PendingChoiceView;

const board = (ids: string[]): Pick<GameView, "battlefield"> =>
  ({ battlefield: { kind: "battlefield", count: ids.length, cards: ids.map(land) } }) as Pick<
    GameView,
    "battlefield"
  >;

beforeEach(() => {
  listFallback.set(new Set());
  targeting.set(null);
});

describe("answeredOnBoard", () => {
  it("sends a choose_cards over battlefield permanents to the board", () => {
    expect(answeredOnBoard(untapLands(), board(["f1", "f2", "f3", "x"]))).toBe(true);
  });

  it("sends an untap_choice over battlefield permanents to the board", () => {
    expect(answeredOnBoard(untapLands({ kind: "untap_choice" }), board(["f1", "f2", "f3"]))).toBe(
      true,
    );
  });

  it("keeps the grid when any candidate is not on the battlefield", () => {
    expect(answeredOnBoard(untapLands(), board(["f1", "f2"]))).toBe(false);
  });

  it("keeps the grid for a choose_cards with no candidates", () => {
    expect(answeredOnBoard(untapLands({ options: [] }), board([]))).toBe(false);
  });

  it("keeps the grid for other card-set kinds", () => {
    expect(answeredOnBoard(untapLands({ kind: "reveal_pick" }), board(["f1", "f2", "f3"]))).toBe(
      false,
    );
  });

  it("hands the choice back to the grid once the player asks for the list", () => {
    showChoiceAsList("untap-1");
    expect(get(listFallback).has("untap-1")).toBe(true);
    expect(answeredOnBoard(untapLands(), board(["f1", "f2", "f3"]))).toBe(false);
    // Another choice is not affected.
    expect(answeredOnBoard(untapLands({ id: "untap-2" }), board(["f1", "f2", "f3"]))).toBe(true);
  });

  it("still answers the original board kinds on the board", () => {
    expect(answeredOnBoard({ id: "l", kind: "legend_rule" } as PendingChoiceView, board([]))).toBe(
      true,
    );
  });
});

describe("the card-set targeting state", () => {
  it("lights the candidates with the choice's bounds", () => {
    const t = cardSetPickState(untapLands(), finale);
    expect(isCardSetPick(t)).toBe(true);
    expect([...(t.legal?.cards ?? [])]).toEqual(["f1", "f2", "f3"]);
    expect(t.legal?.players.size).toBe(0);
    expect(t.min).toBe(0);
    expect(t.max).toBe(5);
    expect(isMultiPick(t)).toBe(true);
    // Up to five, or none: Done is live with nothing picked.
    expect(canConfirm(t)).toBe(true);
  });

  it("is a multi-pick even at one, so choosing none stays possible", () => {
    const t = cardSetPickState(untapLands({ choose_max: 1 }), finale);
    expect(isMultiPick(t)).toBe(true);
  });

  it("reads a missing ceiling as every candidate", () => {
    const t = cardSetPickState(untapLands({ choose_max: 0 }), finale);
    expect(t.max).toBe(3);
  });

  it("opens through beginChoice", () => {
    beginChoice(untapLands(), finale);
    const t = get(targeting);
    expect(t?.choiceID).toBe("untap-1");
    expect(t && isCardSetPick(t)).toBe(true);
  });

  it("answers with card_ids, the grid's payload", () => {
    let t = cardSetPickState(untapLands(), finale);
    t = togglePick(t, { kind: "card", id: "f2" });
    t = togglePick(t, { kind: "card", id: "f3" });
    expect(choiceAnswer(t, t.picked)).toEqual({ choice_id: "untap-1", card_ids: ["f2", "f3"] });
  });

  it("leaves a target prompt's answer as targets", () => {
    beginChoice(
      {
        id: "pt-1",
        kind: "pick_target",
        chooser: "me",
        pick_target: { cards: ["f1"], players: [] },
      } as unknown as PendingChoiceView,
      finale,
    );
    const t = get(targeting)!;
    const refs = [{ kind: "card" as const, id: "f1" }];
    expect(choiceAnswer(t, refs)).toEqual({ choice_id: "pt-1", targets: refs });
  });
});

describe("the card-set dock request", () => {
  it("asks on the board, keeps Done, and offers the list", () => {
    const t = cardSetPickState(untapLands(), finale);
    const shown: string[] = [];
    const req = targetingRequest(t, null, {
      onDone: () => {},
      onCancel: () => {},
      onShowList: (id) => shown.push(id),
    });
    expect(req.label).toBe(L.chooseOnBoard("Finale of Revelation"));
    expect(req.question).toBe(
      "Finale of Revelation — untap up to five lands — click up to 5 of the highlighted permanents",
    );
    expect(req.detail).toContain("3 to choose from");
    expect(req.detail).toContain("0/5 picked");
    expect(req.primary?.label).toBe("Done");
    expect(req.primary?.disabled).toBe(false);
    expect((req.secondary ?? []).map((a) => a.label)).toEqual([L.showAsList]);
    req.secondary?.[0].onPress();
    expect(shown).toEqual(["untap-1"]);
  });

  it("says how many when the count is exact", () => {
    const t = cardSetPickState(untapLands({ choose_min: 2, choose_max: 2 }), finale);
    const req = targetingRequest(t, null, { onDone: () => {}, onCancel: () => {} });
    expect(req.question).toContain("click 2 of the highlighted permanents");
    expect(req.primary?.disabled).toBe(true);
  });
});
