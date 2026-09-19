import { describe, expect, it } from "vitest";
import { deckSubtitle, isSignedIn, ZERO_USER_ID, type MyDeckInfo } from "./myDecks";

function deck(partial: Partial<MyDeckInfo>): MyDeckInfo {
  return {
    id: "d1",
    name: "Deck",
    commanders: [],
    card_count: 0,
    updated_at: "2026-09-19T08:00:00Z",
    ...partial,
  };
}

describe("isSignedIn", () => {
  it("is false for undefined (no user_id on the wire at all)", () => {
    expect(isSignedIn(undefined)).toBe(false);
  });

  it("is false for the empty string", () => {
    expect(isSignedIn("")).toBe(false);
  });

  // The bug this function exists to avoid: Principal.user_id's
  // omitempty tag does not actually omit a zero uuid.UUID (it's a
  // fixed-length array, never "empty" to encoding/json), so a guest's
  // or admin's session still carries this literal string. A naive
  // `!!user_id` check would show the picker to every guest.
  it("is false for the zero uuid the server actually sends for a guest", () => {
    expect(isSignedIn(ZERO_USER_ID)).toBe(false);
    expect(isSignedIn("00000000-0000-0000-0000-000000000000")).toBe(false);
  });

  it("is true for a real id", () => {
    expect(isSignedIn("3fa85f64-5717-4562-b3fc-2c963f66afa6")).toBe(true);
  });
});

describe("deckSubtitle", () => {
  it("joins one commander with the card count", () => {
    expect(deckSubtitle(deck({ commanders: ["Atraxa, Praetors' Voice"], card_count: 100 }))).toBe(
      "Atraxa, Praetors' Voice · 100 cards",
    );
  });

  it("joins partner commanders with a slash", () => {
    const line = deckSubtitle(
      deck({ commanders: ["Tymna the Weaver", "Kraum, Ludevic's Opus"], card_count: 100 }),
    );
    expect(line).toBe("Tymna the Weaver / Kraum, Ludevic's Opus · 100 cards");
  });

  it("singularises a one-card count", () => {
    expect(deckSubtitle(deck({ commanders: ["X"], card_count: 1 }))).toContain("1 card");
    expect(deckSubtitle(deck({ commanders: ["X"], card_count: 1 }))).not.toContain("1 cards");
  });

  it("falls back to just the count with no commanders", () => {
    expect(deckSubtitle(deck({ commanders: [], card_count: 100 }))).toBe("100 cards");
  });

  it("drops blank commander entries", () => {
    expect(deckSubtitle(deck({ commanders: [""], card_count: 100 }))).toBe("100 cards");
  });
});
