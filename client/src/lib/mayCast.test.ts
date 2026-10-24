import { describe, expect, it } from "vitest";
import { keywordCastWords, mayCastCopy } from "./mayCast";
import { freeCastLocation, freeCastTarget, mayCastOfferOpen } from "./freeCastRequest";
import type { CardView, PendingChoiceView } from "./protocol";

describe("mayCastCopy (ADR 0099 §7)", () => {
  it("prefers the server's branch labels", () => {
    const c = mayCastCopy(undefined, "Cast it for {R}", "Put it into your graveyard");
    expect(c.accept).toBe("Cast it for {R}");
    expect(c.decline).toBe("Put it into your graveyard");
  });

  it("falls back to generic copy for an unnamed offer", () => {
    const c = mayCastCopy(undefined);
    expect(c.accept).toBe("Cast it free");
    expect(c.decline).toBe("Don't cast it");
  });

  // #4250: the keyword casts ask a resolution_cast; a may_cast under one of
  // their keywords gets the generic words, not a wording of its own.
  it("no longer words the keyword casts", () => {
    for (const k of ["cascade", "discover", "rebound", "suspend", "madness", "free_or_hand"]) {
      expect(mayCastCopy(k).source).toBe("cast · CR 608.2g");
    }
  });
});

describe("keywordCastWords (#4243, #4250)", () => {
  it("names each keyword's rule", () => {
    expect(keywordCastWords("cascade")?.src).toBe("CR 702.85");
    expect(keywordCastWords("discover")?.src).toBe("CR 701.57");
    expect(keywordCastWords("rebound")?.src).toBe("CR 702.88");
    expect(keywordCastWords("suspend")?.src).toBe("CR 702.62");
    expect(keywordCastWords("madness")?.src).toBe("CR 702.35");
    expect(keywordCastWords("paid_or_else")?.src).toBe("CR 608.2g");
  });

  it("says what each decline does", () => {
    expect(keywordCastWords("cascade")?.ifNot).toContain("bottom of your library");
    expect(keywordCastWords("discover")?.ifNot).toContain("into your hand");
    expect(keywordCastWords("free_or_hand")?.ifNot).toBe("it goes into your hand");
    expect(keywordCastWords("rebound")?.ifNot).toBe("it stays in exile");
    expect(keywordCastWords("suspend")?.ifNot).toBe("it stays in exile");
    expect(keywordCastWords("madness")?.ifNot).toBe("it goes into your graveyard");
  });

  it("words the price: free, madness, or the card's own cost", () => {
    expect(keywordCastWords("rebound")?.how).toBe("without paying its mana cost");
    expect(keywordCastWords("madness")?.how).toBe("for its madness cost");
    expect(keywordCastWords("paid_or_else")?.how).toBe("paying its mana cost");
  });

  // #4250, part 2: an effect's own cast, from the hand, free from elsewhere,
  // or at a price; the decline label says where the card stays.
  it("words the effect casts and where a declined card stays", () => {
    expect(keywordCastWords("hand_cast")).toEqual({
      src: "CR 608.2g",
      how: "without paying its mana cost",
      ifNot: "it stays in your hand",
      many: "a spell from your hand",
    });
    expect(keywordCastWords("free_cast", "Leave it in exile")?.ifNot).toBe("it stays in exile");
    expect(keywordCastWords("free_cast", "Leave it in the graveyard")?.ifNot).toBe(
      "it stays in the graveyard",
    );
    expect(keywordCastWords("free_cast", "Put them on the bottom")?.ifNot).toBe(
      "it goes to the bottom of your library with the rest",
    );
    expect(keywordCastWords("paid_cast", "Leave it in exile")?.how).toBe("paying its cost");
    expect(keywordCastWords("paid_cast", "Leave it in the graveyard")?.ifNot).toBe(
      "it stays in the graveyard",
    );
  });

  it("tells the exile-until shape's two declines apart by the server's label", () => {
    expect(keywordCastWords("exile_until", "Put it on the bottom")?.ifNot).toMatch(
      /^it goes to the bottom/,
    );
    expect(keywordCastWords("exile_until", "Leave it in exile")?.ifNot).toMatch(
      /^it stays in exile/,
    );
  });

  it("is null for a copy's offer, which names no keyword", () => {
    expect(keywordCastWords(undefined)).toBeNull();
    expect(keywordCastWords("")).toBeNull();
  });
});

describe("freeCastTarget", () => {
  const card = (id: string, castable: boolean) =>
    ({ instance_id: id, castable_here: castable }) as unknown as CardView;

  it("waits until the card is in exile", () => {
    expect(freeCastTarget([], "a")).toBeUndefined();
  });

  it("returns the card once it is castable", () => {
    expect(freeCastTarget([card("a", true)], "a")?.instance_id).toBe("a");
  });

  it("drops a request whose card is in exile but not castable", () => {
    expect(freeCastTarget([card("a", false)], "a")).toBeNull();
  });

  it("finds an offered card in the viewer's library when it is not in exile (#3801)", () => {
    expect(freeCastLocation([], [card("a", true)], "a")).toEqual({
      card: card("a", true),
      zone: "library",
    });
    expect(freeCastLocation([card("a", true)], [], "a")?.zone).toBe("exile");
    expect(freeCastLocation([], [card("a", false)], "a")).toBeNull();
    expect(freeCastLocation([], [], "a")).toBeUndefined();
    expect(freeCastLocation([], [], null)).toBeNull();
  });

  // #3490: rebound's card is in exile, uncastable, on the frame "Cast it
  // free" is clicked on. That frame must read "not yet", or the request is
  // dropped before the grant arrives and nothing opens.
  it("waits while the offer is still open, then finds the granted card (#3490)", () => {
    const pending = [
      { id: "p", kind: "may_cast", chooser: "me", may_cast_card: "a" },
    ] as unknown as PendingChoiceView[];
    const open = mayCastOfferOpen(pending, "a", "me");
    expect(open).toBe(true);
    expect(freeCastLocation([card("a", false)], [], "a", open)).toBeUndefined();
    // The answer's frame: the prompt is gone and the card carries the grant.
    expect(mayCastOfferOpen([], "a", "me")).toBe(false);
    expect(freeCastLocation([card("a", true)], [], "a", false)?.zone).toBe("exile");
  });

  it("only counts the viewer's own offer of that card as open (#3490)", () => {
    const pending = [
      { id: "p", kind: "may_cast", chooser: "them", may_cast_card: "a" },
      { id: "q", kind: "may_cast", chooser: "me", may_cast_card: "b" },
      { id: "r", kind: "choose_cards", chooser: "me" },
    ] as unknown as PendingChoiceView[];
    expect(mayCastOfferOpen(pending, "a", "me")).toBe(false);
    expect(mayCastOfferOpen(pending, "b", "me")).toBe(true);
    expect(mayCastOfferOpen(pending, null, "me")).toBe(false);
    expect(mayCastOfferOpen(undefined, "a", "me")).toBe(false);
  });

  it("words a play-or-hand prompt (#3801)", () => {
    const c = mayCastCopy("play_or_hand");
    expect(c.accept).toBe("Play it");
    expect(c.decline).toBe("Put it into your hand");
  });
});
