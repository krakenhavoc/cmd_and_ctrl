// router.test.ts — parseHash coverage for the site portal (#1386):
// #/home and #/roadmap, plus a couple of the existing routes so a
// regression here doesn't silently start returning the default
// (login) route for either of them.

import { describe, expect, it } from "vitest";

import { parseHash } from "./router";

describe("parseHash", () => {
  it("parses #/home", () => {
    expect(parseHash("#/home")).toEqual({ name: "home" });
  });

  it("parses #/roadmap", () => {
    expect(parseHash("#/roadmap")).toEqual({ name: "roadmap" });
  });

  // ADR 0112 §3 item 2: the decks page is #/decks, and #/deck-check
  // parses to the SAME route, permanently. The Discord bot's
  // /c2-deck-check replies link #/deck-check?url=<link> (and a pasted
  // check links #/deck-check plainly), and every reply already posted
  // must keep opening the full report. Do not drop this alias.
  it("parses #/deck-check and #/deck-check?url= as the decks page, permanently", () => {
    expect(parseHash("#/deck-check")).toEqual({ name: "decks" });
    expect(parseHash("#/deck-check?url=https%3A%2F%2Fmoxfield.com%2Fdecks%2FAbC123")).toEqual({
      name: "decks",
      url: "https://moxfield.com/decks/AbC123",
    });
    // An empty ?url= is the same as none — no half-populated route.
    expect(parseHash("#/deck-check?url=")).toEqual({ name: "decks" });
    expect(parseHash("#/deck-check")).toEqual(parseHash("#/decks"));
    expect(parseHash("#/deck-check?url=x")).toEqual(parseHash("#/decks?url=x"));
  });

  it("parses #/practice, the tutorial's practice table (ADR 0076)", () => {
    expect(parseHash("#/practice")).toEqual({ name: "practice" });
  });

  it("still parses the pre-existing routes", () => {
    expect(parseHash("#/login")).toEqual({ name: "login" });
    expect(parseHash("#/lobby")).toEqual({ name: "lobby" });
    expect(parseHash("#/catalog")).toEqual({ name: "catalog" });
    expect(parseHash("#/my-games")).toEqual({ name: "myGames" });
  });

  it("parses #/decks, the one decks page (ADR 0112 §3), with and without ?url=", () => {
    expect(parseHash("#/decks")).toEqual({ name: "decks" });
    expect(parseHash("#/decks?url=https%3A%2F%2Farchidekt.com%2Fdecks%2F42")).toEqual({
      name: "decks",
      url: "https://archidekt.com/decks/42",
    });
    expect(parseHash("#/decks?url=")).toEqual({ name: "decks" });
  });

  it("carries the catalogue's ?q= search, decoded", () => {
    expect(parseHash("#/catalog?q=Sol%20Ring")).toEqual({ name: "catalog", query: "Sol Ring" });
    expect(parseHash("#/catalog?q=Borrowing%20100%2C000%20Arrows")).toEqual({
      name: "catalog",
      query: "Borrowing 100,000 Arrows",
    });
    expect(parseHash("#/catalog?q=")).toEqual({ name: "catalog" });
  });

  it("falls back to login for an empty or unknown hash", () => {
    expect(parseHash("")).toEqual({ name: "login" });
    expect(parseHash("#/nowhere")).toEqual({ name: "login" });
  });
});
