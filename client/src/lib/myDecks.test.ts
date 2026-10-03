import { afterEach, describe, expect, it, vi } from "vitest";
import { deleteMyDeck, fetchMyDeckCoverage, fetchMyDecks, renameMyDeck, saveMyDeck } from "./api";
import { LobbyApiError, sessionFromOAuth, setSession } from "./session";
import {
  coverageDetail,
  coverageLine,
  deckSubtitle,
  isSignedIn,
  sourceHost,
  ZERO_USER_ID,
  type MyDeckInfo,
} from "./myDecks";

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

const COVERAGE = {
  counts: { manual: 3, unreviewed: 5, caveats: 4, automated: 40, no_effect: 38 },
  unknown: 1,
  as_printed: 78,
  resolved: 90,
};

describe("coverageLine", () => {
  it("says N of M cards play as printed", () => {
    expect(coverageLine(deck({ coverage: COVERAGE }))).toBe("78 of 90 cards play as printed");
  });

  it("is empty with no coverage or nothing resolved, never a wrong number", () => {
    expect(coverageLine(deck({}))).toBe("");
    expect(coverageLine(deck({ coverage: { ...COVERAGE, resolved: 0, as_printed: 0 } }))).toBe("");
  });
});

describe("coverageDetail", () => {
  it("names the rest in the report's words and skips zero buckets", () => {
    expect(coverageDetail(deck({ coverage: COVERAGE }))).toBe(
      "4 simplified · 5 not checked yet · 3 you resolve by hand · 1 not found",
    );
    expect(
      coverageDetail(
        deck({
          coverage: {
            ...COVERAGE,
            unknown: 0,
            counts: { manual: 0, unreviewed: 0, caveats: 0, automated: 1, no_effect: 1 },
          },
        }),
      ),
    ).toBe("");
  });
});

describe("source links", () => {
  it("shows a short host for a saved link-deck", () => {
    const d = deck({ source_url: "https://www.moxfield.com/decks/abc" });
    expect(sourceHost(d.source_url)).toBe("moxfield.com");
    expect(sourceHost(undefined)).toBe("");
  });
});

describe("library requests", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    setSession(null);
  });

  function signIn(): void {
    setSession(sessionFromOAuth({ token: "tok", expiresAt: "2099-01-01T00:00:00Z", userID: "u1" }));
  }

  it("lists decks with their coverage", async () => {
    signIn();
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ decks: [deck({ coverage: COVERAGE })] }),
    });
    vi.stubGlobal("fetch", fetchMock);
    const res = await fetchMyDecks();
    expect(coverageLine(res.decks[0])).toBe("78 of 90 cards play as printed");
    expect(fetchMock.mock.calls[0][0]).toBe("/me/decks");
  });

  it("renames with PATCH, deletes with DELETE and reads the report", async () => {
    signIn();
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ id: "d1", name: "New", commanders: [], card_count: 1, updated_at: "" }),
    });
    vi.stubGlobal("fetch", fetchMock);
    expect((await renameMyDeck("d1", "New")).name).toBe("New");
    await deleteMyDeck("d1");
    await fetchMyDeckCoverage("d1");
    const calls = fetchMock.mock.calls as [string, RequestInit | undefined][];
    expect(calls[0][0]).toBe("/me/decks/d1");
    expect(calls[0][1]?.method).toBe("PATCH");
    expect(calls[0][1]?.body).toBe(JSON.stringify({ name: "New" }));
    expect(calls[1][0]).toBe("/me/decks/d1");
    expect(calls[1][1]?.method).toBe("DELETE");
    expect(calls[2][0]).toBe("/me/decks/d1/coverage");
  });

  // ADR 0112 §3 item 4: POST /me/decks, the decks page's explicit save.
  it("saves a checked link or list with its name, and leaves an empty name to the server", async () => {
    signIn();
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ deck: deck({ name: "Weekend" }), replaced: false }),
    });
    vi.stubGlobal("fetch", fetchMock);
    const res = await saveMyDeck({ url: "https://archidekt.com/decks/42" }, "  Weekend ");
    expect(res).toEqual({ deck: deck({ name: "Weekend" }), replaced: false });
    await saveMyDeck({ text: "1 Sol Ring" }, "   ");
    const calls = fetchMock.mock.calls as [string, RequestInit][];
    expect(calls[0][0]).toBe("/me/decks");
    expect(calls[0][1].method).toBe("POST");
    expect(JSON.parse(calls[0][1].body as string)).toEqual({
      url: "https://archidekt.com/decks/42",
      name: "Weekend",
    });
    expect(JSON.parse(calls[1][1].body as string)).toEqual({ text: "1 Sol Ring" });
  });

  it("surfaces the full-library 409 as the server words it", async () => {
    signIn();
    const body = {
      error: "Your deck library is full (200 decks). Delete one below to save this one.",
      code: "library_full",
    };
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 409,
        statusText: "Conflict",
        clone: () => ({ json: async () => body }),
      }),
    );
    const err = await saveMyDeck({ text: "1 Sol Ring" }, "x").catch((e: unknown) => e);
    expect(err).toBeInstanceOf(LobbyApiError);
    expect((err as LobbyApiError).status).toBe(409);
    expect((err as LobbyApiError).message).toBe(body.error);
  });

  it("surfaces the server's rename-conflict message", async () => {
    signIn();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 409,
        statusText: "Conflict",
        clone: () => ({ json: async () => ({ error: "you already have a deck with that name" }) }),
      }),
    );
    const err = await renameMyDeck("d1", "Dup").catch((e: unknown) => e);
    expect(err).toBeInstanceOf(LobbyApiError);
    expect((err as LobbyApiError).status).toBe(409);
    expect((err as LobbyApiError).message).toBe("you already have a deck with that name");
  });
});
