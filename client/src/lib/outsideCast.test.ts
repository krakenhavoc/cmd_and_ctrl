import { describe, expect, it } from "vitest";
import {
  CHOOSE_CAST_FREE_OUTSIDE,
  CHOOSE_PLAY_OUTSIDE,
  isOutsideCastPick,
  outsideCastReady,
} from "./freeCastRequest";
import { castableFromZone, outsideCastLabel } from "./zoneBrowser.logic";
import { actionableCount, type LegalActions } from "./legalActions";
import type { CardView, GameView, PendingChoiceView } from "./protocol";

// #3955: casting or playing a card from outside the game. The pick (a
// choose_cards over the sideboard, or an outside_collection over the
// database) leaves a permission on each chosen card in the sideboard; the
// client opens the sideboard tab, whose cards carry the cast buttons.

const pick = (kind: string, dest?: string, chooser = "me"): PendingChoiceView =>
  ({ kind, choose_destination: dest, chooser }) as unknown as PendingChoiceView;

const side = (over: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "c1",
    name: "Ulamog's Crusher",
    type_line: "Creature — Eldrazi",
    ...over,
  }) as CardView;

describe("isOutsideCastPick", () => {
  it("is a sideboard or database pick that casts or plays what it names", () => {
    expect(isOutsideCastPick(pick("choose_cards", CHOOSE_CAST_FREE_OUTSIDE))).toBe(true);
    expect(isOutsideCastPick(pick("outside_collection", CHOOSE_CAST_FREE_OUTSIDE))).toBe(true);
    expect(isOutsideCastPick(pick("choose_cards", CHOOSE_PLAY_OUTSIDE))).toBe(true);
  });

  it("is not a plain take from outside the game, or another kind", () => {
    expect(isOutsideCastPick(pick("choose_cards", "outside"))).toBe(false);
    expect(isOutsideCastPick(pick("outside_collection", "outside"))).toBe(false);
    expect(isOutsideCastPick(pick("may_cast", CHOOSE_CAST_FREE_OUTSIDE))).toBe(false);
    expect(isOutsideCastPick(null)).toBe(false);
  });
});

describe("outsideCastReady", () => {
  it("waits while the viewer still owes the pick", () => {
    const open = [pick("choose_cards", CHOOSE_CAST_FREE_OUTSIDE)];
    expect(outsideCastReady(open, [side({ castable_here: true })], "me")).toBeUndefined();
  });

  it("is ready once the pick is answered and a sideboard card is castable", () => {
    expect(outsideCastReady([], [side({ castable_here: true })], "me")).toBe(true);
  });

  it("is stale when nothing was granted", () => {
    expect(outsideCastReady([], [side()], "me")).toBeNull();
    expect(outsideCastReady(undefined, undefined, "me")).toBeNull();
  });

  it("does not wait on another seat's pick", () => {
    const theirs = [pick("choose_cards", CHOOSE_CAST_FREE_OUTSIDE, "opp")];
    expect(outsideCastReady(theirs, [side({ castable_here: true })], "me")).toBe(true);
  });
});

describe("the sideboard's cast button", () => {
  it("shows on a card the server marked castable_here, and only on the sideboard", () => {
    const granted = side({ castable_here: true });
    expect(castableFromZone(granted, "sideboard")).toBe(true);
    expect(castableFromZone(side(), "sideboard")).toBe(false);
    expect(castableFromZone(granted, "exile")).toBe(false);
  });

  it("reads 'cast free' for a {0} grant, 'play' for Wish's and for a land", () => {
    expect(
      outsideCastLabel(side({ exile_play: { cost_override: "{0}" } as CardView["exile_play"] })),
    ).toBe("cast free");
    expect(outsideCastLabel(side({ exile_play: {} as CardView["exile_play"] }))).toBe("play");
    expect(outsideCastLabel(side({ type_line: "Basic Land — Forest" }))).toBe("play");
    expect(outsideCastLabel(side())).toBe("cast");
  });
});

describe("the ready count", () => {
  it("counts a granted sideboard card as an action", () => {
    const legal = {
      known: true,
      isReady: () => true,
      castableFrom: (_: string, zone: string) => zone === "outside",
      kinds: () => ["cast"],
      castIdleHint: () => undefined,
      readyAbilityRefs: () => [],
      readyManaRefs: () => [],
      readySpecialActions: () => [],
      canAttack: () => false,
      blockableAttackers: () => [],
    } as unknown as LegalActions;
    const view = {
      seats: [
        {
          id: "me",
          hand: { cards: [] },
          graveyard: { cards: [] },
          command: { cards: [] },
          library: { cards: [] },
          sideboard: [side({ castable_here: true })],
        },
      ],
      battlefield: { cards: [] },
      exile: { cards: [] },
    } as unknown as GameView;
    expect(actionableCount(legal, view, "me")).toBe(1);
  });
});
