// router.test.ts — parseHash coverage: the front page and the gate page
// (ADR 0156), #/roadmap, plus a couple of the existing routes so a
// regression here doesn't silently start returning the default
// (login) route for either of them.

import { describe, expect, it } from "vitest";

import { parseHash } from "./router";

describe("parseHash", () => {
  // ADR 0156 §4: #/ is the front page and #/login an alias of it (the
  // OAuth error path and the e2e suite use it). #/home still parses, so
  // the router can send old links on (signedInHome.test.ts).
  it("parses #/ and #/login as the front page, and keeps #/home for its redirect", () => {
    expect(parseHash("#/")).toEqual({ name: "login" });
    expect(parseHash("")).toEqual({ name: "login" });
    expect(parseHash("#/login")).toEqual({ name: "login" });
    expect(parseHash("#/home")).toEqual({ name: "home" });
  });

  // ADR 0156 §5: the gate page, and ?retry=1 when Discord did not answer.
  it("parses #/request-access, with and without ?retry=1", () => {
    expect(parseHash("#/request-access")).toEqual({ name: "requestAccess" });
    expect(parseHash("#/request-access?retry=1")).toEqual({ name: "requestAccess", retry: true });
    expect(parseHash("#/request-access?retry=0")).toEqual({ name: "requestAccess" });
  });

  // ADR 0148 / #3103: the Discord callback lands the browser back on the
  // invite's join page with the capacity code. Only that code is read; any
  // other `error` value is ignored, and a plain join route is unchanged.
  it("parses a join link, and the server_full code the Discord callback adds", () => {
    expect(parseHash("#/games/g1/join?t=inv")).toEqual({
      name: "join",
      gameID: "g1",
      inviteToken: "inv",
      spectator: false,
    });
    expect(parseHash("#/games/g1/join?error=server_full&t=inv")).toEqual({
      name: "join",
      gameID: "g1",
      inviteToken: "inv",
      spectator: false,
      serverFull: true,
    });
    expect(parseHash("#/games/g1/join?error=something_else&t=inv")).toEqual(
      parseHash("#/games/g1/join?t=inv"),
    );
  });

  it("parses #/roadmap", () => {
    expect(parseHash("#/roadmap")).toEqual({ name: "roadmap" });
  });

  // #2983: the privacy notice.
  it("parses #/privacy", () => {
    expect(parseHash("#/privacy")).toEqual({ name: "privacy" });
    expect(parseHash("#/about")).toEqual({ name: "about" });
    expect(parseHash("#/guide")).toEqual({ name: "guide" });
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
