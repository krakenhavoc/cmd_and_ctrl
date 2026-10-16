import { describe, expect, it } from "vitest";
import { mayCastCopy } from "./mayCast";
import {
  freeCastLocation,
  freeCastTarget,
  mayCastKeywordsThatOpenACast,
  mayCastOfferOpen,
} from "./freeCastRequest";
import type { CardView, PendingChoiceView } from "./protocol";

describe("mayCastCopy (ADR 0099 §7)", () => {
  it("words a discover prompt around the hand", () => {
    const c = mayCastCopy("discover");
    expect(c.source).toContain("701.57");
    expect(c.decline).toBe("Put it into your hand");
    expect(c.hint).toContain("hand");
  });

  it("words a cascade prompt around the bottom of the library", () => {
    const c = mayCastCopy("cascade");
    expect(c.source).toContain("702.85");
    expect(c.decline).toBe("Put it on the bottom");
  });

  it("words an exile-until prompt around the bottom of the library and opens a free cast", () => {
    const c = mayCastCopy("exile_until");
    expect(c.decline).toBe("Put it on the bottom");
    expect(c.accept).toBe("Cast it free");
    expect(mayCastKeywordsThatOpenACast.has("exile_until")).toBe(true);
  });

  it("words a free-or-hand prompt around the hand and opens a free cast", () => {
    const c = mayCastCopy("free_or_hand");
    expect(c.decline).toBe("Put it into your hand");
    expect(c.accept).toBe("Cast it free");
    expect(mayCastKeywordsThatOpenACast.has("free_or_hand")).toBe(true);
  });

  it("words a paid-or-else prompt as a paid cast and opens the cast chain", () => {
    const c = mayCastCopy("paid_or_else", undefined, "Don't cast it (2 damage to each opponent)");
    expect(c.accept).toBe("Cast it");
    expect(c.decline).toBe("Don't cast it (2 damage to each opponent)");
    expect(c.hint).toContain("paying its mana cost");
    expect(mayCastKeywordsThatOpenACast.has("paid_or_else")).toBe(true);
  });

  it("words a rebound prompt around staying in exile", () => {
    const c = mayCastCopy("rebound");
    expect(c.source).toContain("702.88");
    expect(c.decline).toBe("Leave it in exile");
    expect(mayCastKeywordsThatOpenACast.has("rebound")).toBe(true);
  });

  it("prefers the server's branch labels", () => {
    const c = mayCastCopy("madness", "Cast it for {R}", "Put it into your graveyard");
    expect(c.accept).toBe("Cast it for {R}");
    expect(c.source).toContain("702.35");
  });

  it("falls back to generic copy for an unnamed offer", () => {
    const c = mayCastCopy(undefined);
    expect(c.accept).toBe("Cast it free");
    expect(c.decline).toBe("Don't cast it");
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

  it("words a play-or-hand prompt and does not open a cast (#3801)", () => {
    const c = mayCastCopy("play_or_hand");
    expect(c.accept).toBe("Play it");
    expect(c.decline).toBe("Put it into your hand");
    expect(mayCastKeywordsThatOpenACast.has("play_or_hand")).toBe(false);
    expect(mayCastKeywordsThatOpenACast.has("free_or_hand")).toBe(true);
  });

  it("only discover, cascade and suspend open a free cast", () => {
    expect(mayCastKeywordsThatOpenACast.has("discover")).toBe(true);
    expect(mayCastKeywordsThatOpenACast.has("madness")).toBe(false);
  });
});
