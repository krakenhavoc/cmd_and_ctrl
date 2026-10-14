import { describe, expect, it } from "vitest";
import { mayCastCopy } from "./mayCast";
import { freeCastTarget, mayCastKeywordsThatOpenACast } from "./freeCastRequest";
import type { CardView } from "./protocol";

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

  it("only discover, cascade and suspend open a free cast", () => {
    expect(mayCastKeywordsThatOpenACast.has("discover")).toBe(true);
    expect(mayCastKeywordsThatOpenACast.has("madness")).toBe(false);
  });
});
