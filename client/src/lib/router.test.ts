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
  it("parses #/linked-role with each status the server sends, and anything else as failed", () => {
    for (const status of ["linked", "signin", "unavailable", "failed"] as const) {
      expect(parseHash(`#/linked-role?status=${status}`)).toEqual({ name: "linkedRole", status });
    }
    expect(parseHash("#/linked-role")).toEqual({ name: "linkedRole", status: "failed" });
    expect(parseHash("#/linked-role?status=admin")).toEqual({
      name: "linkedRole",
      status: "failed",
    });
  });

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

describe("pods routes (ADR 0220 §11)", () => {
  it("parses #/pods, a pod and an invite link", () => {
    expect(parseHash("#/pods")).toEqual({ name: "pods" });
    expect(parseHash("#/pods/abc-123")).toEqual({ name: "pod", id: "abc-123" });
    expect(parseHash("#/pods/abc-123/stats")).toEqual({ name: "podStats", id: "abc-123" });
    expect(parseHash("#/pods/join/tok_en-1")).toEqual({ name: "podJoin", token: "tok_en-1" });
    expect(parseHash("#/pods/join")).toEqual({ name: "pods" });
  });
});

describe("draft routes (ADR 0222 §5, #5438)", () => {
  it("keeps #/draft/sets, the sets page", () => {
    expect(parseHash("#/draft/sets")).toEqual({ name: "draftSets" });
  });

  it("parses the create form, with and without a chosen set", () => {
    expect(parseHash("#/draft")).toEqual({ name: "draftNew" });
    expect(parseHash("#/draft/new")).toEqual({ name: "draftNew" });
    expect(parseHash("#/draft?set=blb")).toEqual({ name: "draftNew", set: "blb" });
    expect(parseHash("#/draft/new?set=blb")).toEqual({ name: "draftNew", set: "blb" });
  });

  it("parses a draft, and the invite on a shared link", () => {
    const id = "6f1c2a9e-0000-4000-8000-000000000000";
    expect(parseHash(`#/draft/${id}`)).toEqual({ name: "draft", id });
    expect(parseHash(`#/draft/${id}?invite=AbC-123`)).toEqual({
      name: "draft",
      id,
      invite: "AbC-123",
    });
    // An empty invite is no invite: the page then looks for a stored seat.
    expect(parseHash(`#/draft/${id}?invite=`)).toEqual({ name: "draft", id });
  });

  it("opens both to a signed-out visitor", async () => {
    const { isPublicRoute, routeRedirect } = await import("./signedInHome");
    expect(isPublicRoute({ name: "draftNew" })).toBe(true);
    expect(isPublicRoute({ name: "draft", id: "x" })).toBe(true);
    expect(routeRedirect({ name: "draftNew" }, null)).toBeNull();
    expect(routeRedirect({ name: "draft", id: "x", invite: "y" }, null)).toBeNull();
  });
});
