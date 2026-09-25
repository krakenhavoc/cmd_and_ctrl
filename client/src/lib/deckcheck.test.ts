import { afterEach, describe, expect, it, vi } from "vitest";

import { currentSession, sessionFromOAuth, setSession } from "./session";
import {
  BUCKET_LABELS,
  BUCKET_ORDER,
  DeckCheckError,
  canRequestCards,
  checkDeck,
  deckRequestOutcomeMessage,
  formatRetryAfter,
  groupCardsByBucket,
  requestDeck,
  shouldOfferPaste,
  type CoverageCard,
  type CoverageReport,
} from "./deckcheck";

// deckcheck.test.ts covers the pure half of #1631 / ADR 0095 §5's
// #/deck-check page: grouping, labelling and error-shape decisions a
// user can get wrong, plus the two fetch wrappers' status-code
// branching (which response shapes are errors vs. normal outcomes).

afterEach(() => {
  vi.unstubAllGlobals();
  setSession(null);
});

function card(over: Partial<CoverageCard> & Pick<CoverageCard, "name" | "bucket">): CoverageCard {
  return { oracle_id: over.name, count: 1, ...over };
}

function report(over: Partial<CoverageReport> = {}): CoverageReport {
  return {
    deck_name: "Needy Deck",
    source: "moxfield",
    source_url: "https://moxfield.com/decks/AbC123",
    deck_key: "moxfield:AbC123",
    commanders: [],
    counts: { manual: 0, unreviewed: 0, caveats: 0, automated: 0, no_effect: 0 },
    cards: [],
    unknown: [],
    violations: [],
    ...over,
  };
}

describe("groupCardsByBucket", () => {
  it("splits cards into every bucket, preserving the given order within each", () => {
    const cards = [
      card({ name: "Doubling Season", bucket: "manual" }),
      card({ name: "Forest", bucket: "no_effect" }),
      card({ name: "Abzan Charm", bucket: "caveats", caveats: ["…"] }),
      card({ name: "Sol Ring", bucket: "automated" }),
      card({ name: "Arcane Denial", bucket: "unreviewed" }),
      card({ name: "Command Tower", bucket: "no_effect" }),
    ];
    const got = groupCardsByBucket(cards);
    expect(got.manual.map((c) => c.name)).toEqual(["Doubling Season"]);
    expect(got.no_effect.map((c) => c.name)).toEqual(["Forest", "Command Tower"]);
    expect(got.caveats.map((c) => c.name)).toEqual(["Abzan Charm"]);
    expect(got.automated.map((c) => c.name)).toEqual(["Sol Ring"]);
    expect(got.unreviewed.map((c) => c.name)).toEqual(["Arcane Denial"]);
  });

  it("returns every bucket, empty, for an empty deck", () => {
    const got = groupCardsByBucket([]);
    for (const b of BUCKET_ORDER) expect(got[b]).toEqual([]);
  });
});

describe("BUCKET_LABELS / BUCKET_ORDER", () => {
  it("has a player-facing label for every bucket, in the server's order", () => {
    expect(BUCKET_ORDER).toEqual(["manual", "unreviewed", "caveats", "automated", "no_effect"]);
    for (const b of BUCKET_ORDER) {
      expect(BUCKET_LABELS[b]).toBeTruthy();
      // Player words, not engine vocabulary — no bare bucket name.
      expect(BUCKET_LABELS[b].toLowerCase()).not.toBe(b);
    }
  });
});

describe("canRequestCards", () => {
  it("is false with no report", () => {
    expect(canRequestCards(null)).toBe(false);
    expect(canRequestCards(undefined)).toBe(false);
  });

  it("is true for a pasted-list check with a manual card (ADR 0095 amendment)", () => {
    const r = report({
      source: "text",
      source_url: undefined,
      deck_key: "list:0123456789abcdef",
      counts: { manual: 3, unreviewed: 0, caveats: 0, automated: 0, no_effect: 0 },
    });
    expect(canRequestCards(r)).toBe(true);
  });

  it("is false for a pasted-list check with nothing to add", () => {
    const r = report({
      source: "text",
      source_url: undefined,
      deck_key: "list:0123456789abcdef",
      counts: { manual: 0, unreviewed: 0, caveats: 2, automated: 5, no_effect: 3 },
    });
    expect(canRequestCards(r)).toBe(false);
  });

  it("is true for a link check with a manual card", () => {
    const r = report({
      counts: { manual: 1, unreviewed: 0, caveats: 0, automated: 40, no_effect: 20 },
    });
    expect(canRequestCards(r)).toBe(true);
  });

  it("is true for a link check with only an unreviewed card", () => {
    const r = report({
      counts: { manual: 0, unreviewed: 1, caveats: 0, automated: 40, no_effect: 20 },
    });
    expect(canRequestCards(r)).toBe(true);
  });

  it("is false for a link check that is fully automated or caveated", () => {
    const r = report({
      counts: { manual: 0, unreviewed: 0, caveats: 6, automated: 40, no_effect: 20 },
    });
    expect(canRequestCards(r)).toBe(false);
  });
});

describe("formatRetryAfter", () => {
  it("renders seconds under a minute", () => {
    expect(formatRetryAfter(0)).toBe("0 seconds");
    expect(formatRetryAfter(1)).toBe("1 second");
    expect(formatRetryAfter(45)).toBe("45 seconds");
  });

  it("renders minutes once it rounds up to at least one", () => {
    expect(formatRetryAfter(60)).toBe("1 minute");
    expect(formatRetryAfter(90)).toBe("2 minutes");
    expect(formatRetryAfter(150)).toBe("3 minutes");
  });

  it("renders hours once it rounds up to at least one", () => {
    expect(formatRetryAfter(3600)).toBe("1 hour");
    expect(formatRetryAfter(7200)).toBe("2 hours");
  });
});

describe("deckRequestOutcomeMessage", () => {
  it("distinguishes joined from already-requested", () => {
    expect(deckRequestOutcomeMessage({ status: "joined" })).toMatch(/added your name/i);
    expect(deckRequestOutcomeMessage({ status: "joined", already_requested: true })).toMatch(
      /already asked/i,
    );
  });

  it("names filed, nothing-to-add and rate-limited outcomes", () => {
    expect(deckRequestOutcomeMessage({ status: "filed" })).toMatch(/filed/i);
    expect(deckRequestOutcomeMessage({ status: "nothing_to_add" })).toMatch(/nothing to add/i);
    expect(deckRequestOutcomeMessage({ status: "rate_limited", retry_after: 90 })).toBe(
      "You've asked a few times already — try again in 2 minutes.",
    );
  });
});

describe("checkDeck", () => {
  it("returns the parsed report on 200", async () => {
    const body = report({ deck_name: "Atraxa Superfriends" });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, json: async () => body }));
    await expect(checkDeck({ url: "https://moxfield.com/decks/AbC123" })).resolves.toEqual(body);
  });

  it("throws the server's sentence, code and violations on a fetch error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 422,
        statusText: "Unprocessable Entity",
        json: async () => ({
          error: "That deck is private. Make it public or unlisted, or paste the list as text.",
          code: "deck_private",
          violations: [{ code: "deck_private", card: "<the url>", message: "…" }],
        }),
      }),
    );
    const err = await checkDeck({ url: "x" }).catch((e: unknown) => e);
    expect(err).toBeInstanceOf(DeckCheckError);
    const dce = err as DeckCheckError;
    expect(dce.status).toBe(422);
    expect(dce.code).toBe("deck_private");
    expect(dce.message).toMatch(/private/);
    expect(dce.violations).toHaveLength(1);
  });

  it("carries the paste_list hint on a Moxfield error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 502,
        statusText: "Bad Gateway",
        json: async () => ({
          error:
            "Moxfield blocks our server. On Moxfield, open the deck → Export → Copy plain text, then paste the list here instead.",
          code: "upstream_blocked",
          hint: "paste_list",
          violations: [{ code: "upstream_blocked", card: "<the url>", message: "…" }],
        }),
      }),
    );
    const err = await checkDeck({ url: "https://moxfield.com/decks/x" }).catch((e: unknown) => e);
    expect(err).toBeInstanceOf(DeckCheckError);
    expect((err as DeckCheckError).code).toBe("upstream_blocked");
    expect((err as DeckCheckError).message).toMatch(/Moxfield blocks our server/);
    expect(shouldOfferPaste(err)).toBe(true);
  });

  it("offers no paste on an error without the hint", () => {
    expect(shouldOfferPaste(new DeckCheckError(502, "x", "upstream_blocked"))).toBe(false);
    expect(shouldOfferPaste(new Error("x"))).toBe(false);
  });

  it("rewrites a bare 429 (no status field) to the friendly sentence", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 429,
        statusText: "Too Many Requests",
        json: async () => ({ error: "too many requests" }),
      }),
    );
    const err = await checkDeck({ text: "1 Sol Ring" }).catch((e: unknown) => e);
    expect(err).toBeInstanceOf(DeckCheckError);
    expect((err as DeckCheckError).message).toBe("Too many checks — try again in a few seconds.");
  });
});

describe("requestDeck", () => {
  it("sends the session's bearer token and returns a filed outcome", async () => {
    setSession(
      sessionFromOAuth({ token: "tok-123", expiresAt: "2099-01-01T00:00:00Z", userID: "u1" }),
    );
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({
        status: "filed",
        issue_url: "https://github.com/krakenhavoc/cmd_and_ctrl/issues/1700",
        issue_number: 1700,
        report: report(),
      }),
    });
    vi.stubGlobal("fetch", fetchMock);
    const res = await requestDeck({ url: "https://moxfield.com/decks/AbC123" });
    expect(res.status).toBe("filed");
    expect(res.issue_number).toBe(1700);
    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    const headers = init.headers as Headers;
    expect(headers.get("Authorization")).toBe("Bearer tok-123");
  });

  it("sends a pasted list as {text}", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ status: "filed", issue_number: 1701, report: report() }),
    });
    vi.stubGlobal("fetch", fetchMock);
    await requestDeck({ text: "1 Sol Ring" });
    const [path, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(path).toBe("/deck-requests");
    expect(JSON.parse(init.body as string)).toEqual({ text: "1 Sol Ring" });
  });

  it("returns rate_limited as a normal outcome, not a throw", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 429,
        json: async () => ({ status: "rate_limited", retry_after: 3600 }),
      }),
    );
    const res = await requestDeck({ url: "https://moxfield.com/decks/AbC123" });
    expect(res).toEqual({ status: "rate_limited", retry_after: 3600 });
  });

  it("throws on a bare 429 with no status field", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 429,
        statusText: "Too Many Requests",
        json: async () => ({ error: "too many requests" }),
      }),
    );
    await expect(requestDeck({ url: "https://moxfield.com/decks/AbC123" })).rejects.toThrow(
      "Too many checks — try again in a few seconds.",
    );
  });

  it("clears the session on 401, the server's own word the credential is gone", async () => {
    setSession(
      sessionFromOAuth({ token: "stale", expiresAt: "2099-01-01T00:00:00Z", userID: "u1" }),
    );
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        statusText: "Unauthorized",
        json: async () => ({ error: "session expired" }),
      }),
    );
    await expect(requestDeck({ url: "https://moxfield.com/decks/AbC123" })).rejects.toBeInstanceOf(
      DeckCheckError,
    );
    expect(currentSession()).toBeNull();
  });

  it("throws with the server's sentence on 403 (no Discord identity)", async () => {
    setSession(sessionFromOAuth({ token: "tok", expiresAt: "2099-01-01T00:00:00Z" }));
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        statusText: "Forbidden",
        json: async () => ({ error: "your account has no Discord identity" }),
      }),
    );
    const err = await requestDeck({ url: "https://moxfield.com/decks/AbC123" }).catch(
      (e: unknown) => e,
    );
    expect(err).toBeInstanceOf(DeckCheckError);
    expect((err as DeckCheckError).status).toBe(403);
    expect((err as DeckCheckError).message).toMatch(/discord identity/i);
  });
});
