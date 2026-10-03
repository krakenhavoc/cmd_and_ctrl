// @vitest-environment jsdom
//
// decksPage.test.ts — the pure half of ADR 0112 §3, the one decks page:
// who may request, save and see a library; the report's "as printed"
// line; the save button's name and label; and the sign-in return route
// (§3 item 7), which must never send the browser anywhere but #/decks.

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import type { CoverageReport } from "./deckcheck";
import {
  AFTER_SIGN_IN_KEY,
  AFTER_SIGN_IN_TEXT_KEY,
  decksAccess,
  defaultSaveName,
  isDecksReturn,
  libraryDeckRequestable,
  rememberAfterSignIn,
  reportAsPrinted,
  returnHashFor,
  saveButtonLabel,
  savedMessage,
  takeAfterSignIn,
  takePendingDeckText,
} from "./decksPage";
import type { MyDeckInfo } from "./myDecks";
import type { Session } from "./session";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const NIL = "00000000-0000-0000-0000-000000000000";

function sess(role: Session["principal"]["role"], userID?: string): Session {
  return {
    token: "t",
    expiresAt: "2099-01-01T00:00:00Z",
    principal: {
      role,
      user_id: userID,
      name: "Alice",
      issued_at: "2026-01-01T00:00:00Z",
      expires_at: "2099-01-01T00:00:00Z",
    },
    gameID: role === "player" || role === "spectator" ? "g1" : undefined,
  };
}

function report(over: Partial<CoverageReport> = {}): CoverageReport {
  return {
    deck_name: "Weekend deck",
    source: "archidekt",
    source_url: "https://archidekt.com/decks/42",
    commanders: ["Atraxa, Praetors' Voice"],
    counts: { manual: 2, unreviewed: 1, caveats: 3, automated: 50, no_effect: 10 },
    cards: [],
    unknown: [],
    violations: [],
    ...over,
  };
}

function libDeck(over: Partial<MyDeckInfo> = {}): MyDeckInfo {
  return {
    id: "d1",
    name: "Atraxa",
    commanders: [],
    card_count: 100,
    updated_at: "2026-09-19T08:00:00Z",
    ...over,
  };
}

describe("decksAccess (§3 item 3, who can do what)", () => {
  it("lets a signed-in person, in either mode, do everything", () => {
    for (const s of [sess("identified", USER), sess("player", USER), sess("spectator", USER)]) {
      expect(decksAccess(s)).toEqual({ request: "allowed", save: "allowed", library: "allowed" });
    }
    const admin = { ...sess("player", USER), admin: true };
    expect(decksAccess(admin)).toEqual({ request: "allowed", save: "allowed", library: "allowed" });
  });

  it("asks a signed-out visitor to sign in with Discord", () => {
    expect(decksAccess(null)).toEqual({ request: "signIn", save: "signIn", library: "signIn" });
  });

  it("tells a guest seat or spectator to link Discord at their table, never to sign in", () => {
    for (const s of [sess("player", NIL), sess("player"), sess("spectator")]) {
      expect(decksAccess(s)).toEqual({
        request: "linkDiscord",
        save: "linkDiscord",
        library: "linkDiscord",
      });
    }
  });

  it("hides request, save and the library from the admin token", () => {
    expect(decksAccess(sess("admin"))).toEqual({
      request: "hidden",
      save: "hidden",
      library: "hidden",
    });
  });

  it("lets a Discord sign-in with no user database request, but not save", () => {
    expect(decksAccess(sess("identified"))).toEqual({
      request: "allowed",
      save: "hidden",
      library: "hidden",
    });
  });
});

describe("reportAsPrinted (§3 item 1.2)", () => {
  it("counts automated plus nothing-to-automate over every bucketed card", () => {
    expect(reportAsPrinted(report())).toBe("60 of 66 cards play as printed");
  });

  it("says nothing for an empty report", () => {
    const counts = { manual: 0, unreviewed: 0, caveats: 0, automated: 0, no_effect: 0 };
    expect(reportAsPrinted(report({ counts }))).toBe("");
  });
});

describe("the save field and button (§3 item 4)", () => {
  it("fills the name with the deck's name, then its first commander", () => {
    expect(defaultSaveName(report())).toBe("Weekend deck");
    expect(defaultSaveName(report({ deck_name: "  " }))).toBe("Atraxa, Praetors' Voice");
    expect(defaultSaveName(report({ deck_name: "", commanders: [] }))).toBe("");
  });

  it("reads Replace ‹name› before anything is sent when the name is taken", () => {
    const lib = [libDeck({ name: "Atraxa" })];
    expect(saveButtonLabel("Atraxa", lib)).toBe("Replace Atraxa");
    expect(saveButtonLabel("  Atraxa  ", lib)).toBe("Replace Atraxa");
    // The server matches names exactly.
    expect(saveButtonLabel("atraxa", lib)).toBe("Save to my decks");
    expect(saveButtonLabel("", lib)).toBe("Save to my decks");
  });

  it("says whether the save made a deck or replaced one", () => {
    expect(savedMessage({ deck: libDeck({ name: "A" }), replaced: false })).toBe(
      "Saved A to your decks.",
    );
    expect(savedMessage({ deck: libDeck({ name: "A" }), replaced: true })).toBe(
      "Replaced A in your decks.",
    );
  });
});

describe("libraryDeckRequestable (§3 item 5)", () => {
  it("offers the request on a pasted deck as well as a link deck", () => {
    const counts = { manual: 1, unreviewed: 0, caveats: 0, automated: 5, no_effect: 1 };
    const coverage = { counts, unknown: 0, as_printed: 6, resolved: 7 };
    expect(libraryDeckRequestable(libDeck({ coverage }))).toBe(true);
    expect(libraryDeckRequestable(libDeck({ coverage, source_url: "https://x" }))).toBe(true);
  });

  it("offers it when the coverage is unknown, and not when nothing is missing", () => {
    expect(libraryDeckRequestable(libDeck({}))).toBe(true);
    const counts = { manual: 0, unreviewed: 0, caveats: 2, automated: 5, no_effect: 1 };
    expect(
      libraryDeckRequestable(
        libDeck({ coverage: { counts, unknown: 0, as_printed: 6, resolved: 8 } }),
      ),
    ).toBe(false);
  });
});

describe("the sign-in return route (§3 item 7)", () => {
  beforeEach(() => sessionStorage.clear());
  afterEach(() => sessionStorage.clear());

  it("accepts only the decks route", () => {
    expect(isDecksReturn("#/decks")).toBe(true);
    expect(isDecksReturn("#/decks?url=https%3A%2F%2Farchidekt.com%2Fdecks%2F42")).toBe(true);
    expect(isDecksReturn("#/deck-check")).toBe(true);
    for (const bad of [
      "",
      "#/lobby",
      "#/games/g1",
      "#/nowhere",
      "https://evil.example/#/decks",
      "//evil.example/#/decks",
      "javascript:alert(1)",
      null,
      undefined,
    ]) {
      expect(isDecksReturn(bad)).toBe(false);
    }
  });

  it("returns to the checked link, or to the page for a paste", () => {
    expect(returnHashFor({ url: "https://archidekt.com/decks/42" })).toBe(
      "#/decks?url=" + encodeURIComponent("https://archidekt.com/decks/42"),
    );
    expect(returnHashFor({ text: "1 Sol Ring" })).toBe("#/decks");
    expect(returnHashFor(null)).toBe("#/decks");
  });

  it("stores the route and a pasted list, and hands both back once", () => {
    rememberAfterSignIn({ text: "1 Sol Ring" });
    expect(sessionStorage.getItem(AFTER_SIGN_IN_KEY)).toBe("#/decks");
    expect(sessionStorage.getItem(AFTER_SIGN_IN_TEXT_KEY)).toBe("1 Sol Ring");

    expect(takeAfterSignIn()).toBe("#/decks");
    expect(sessionStorage.getItem(AFTER_SIGN_IN_KEY)).toBeNull();
    expect(sessionStorage.getItem(AFTER_SIGN_IN_TEXT_KEY)).toBeNull();
    expect(takePendingDeckText()).toBe("1 Sol Ring");
    expect(takePendingDeckText()).toBe("");
    expect(takeAfterSignIn()).toBeNull();
  });

  it("stores a checked link as the route, with no list", () => {
    rememberAfterSignIn({ url: "https://archidekt.com/decks/42" });
    expect(sessionStorage.getItem(AFTER_SIGN_IN_TEXT_KEY)).toBeNull();
    expect(takeAfterSignIn()).toBe(
      "#/decks?url=" + encodeURIComponent("https://archidekt.com/decks/42"),
    );
    expect(takePendingDeckText()).toBe("");
  });

  it("drops a stored value that is not the decks route, and its list with it", () => {
    sessionStorage.setItem(AFTER_SIGN_IN_KEY, "https://evil.example/");
    sessionStorage.setItem(AFTER_SIGN_IN_TEXT_KEY, "1 Sol Ring");
    expect(takeAfterSignIn()).toBeNull();
    expect(sessionStorage.getItem(AFTER_SIGN_IN_KEY)).toBeNull();
    expect(sessionStorage.getItem(AFTER_SIGN_IN_TEXT_KEY)).toBeNull();
    expect(takePendingDeckText()).toBe("");
  });
});
