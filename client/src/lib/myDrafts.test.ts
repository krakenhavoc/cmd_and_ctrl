// @vitest-environment jsdom
//
// #5440 (ADR 0222 §9): a signed-in person's finished drafts.

import { afterEach, describe, expect, it, vi } from "vitest";

import {
  basicsText,
  draftDate,
  draftRecord,
  fetchMyDraft,
  fetchMyDrafts,
  myDraftHash,
  opponentText,
  roundText,
  splitPicks,
  type MyDraftMatch,
} from "./myDrafts";
import { parseHash } from "./router";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

const match = (over: Partial<MyDraftMatch>): MyDraftMatch => ({
  round: 1,
  opponent_kind: "person",
  result: "win",
  games_won: 2,
  games_lost: 1,
  ...over,
});

describe("the words for a stored draft", () => {
  it("names an opponent by kind, and by name only when there is one", () => {
    expect(opponentText(match({ opponent_name: "Bea" }))).toBe("Bea");
    expect(opponentText(match({}))).toBe("a player");
    expect(opponentText(match({ opponent_kind: "guest" }))).toBe("a guest");
    expect(opponentText(match({ opponent_kind: "bot" }))).toBe("a bot");
  });

  it("writes each round as a sentence", () => {
    expect(roundText(match({ opponent_name: "Bea" }))).toBe("Won 2–1 against Bea");
    expect(
      roundText(match({ opponent_kind: "bot", result: "loss", games_won: 0, games_lost: 1 })),
    ).toBe("Lost 0–1 to a bot");
    expect(roundText(match({ result: "draw", games_won: 1, games_lost: 1 }))).toBe(
      "Drew 1–1 with a player",
    );
    expect(roundText(match({ opponent_kind: "bye", result: "" }))).toBe("Sat the round out");
  });

  it("writes the record, the basics and the date", () => {
    expect(draftRecord({ wins: 2, losses: 1, draws: 0 })).toBe("2–1");
    expect(draftRecord({ wins: 1, losses: 1, draws: 1 })).toBe("1–1–1");
    expect(basicsText({ Forest: 9, Island: 8, Swamp: 0 })).toBe("8 Island, 9 Forest");
    expect(basicsText(undefined)).toBe("");
    expect(draftDate("2026-11-11T19:02:11Z", "en-GB")).toBe("11 Nov 2026");
    expect(draftDate("not a date")).toBe("");
  });

  it("splits the picks into the deck and the sideboard, each with an id to key on", () => {
    const { deck, side } = splitPicks({
      picks: [
        { name: "A", in_deck: true },
        { name: "B", in_deck: false },
        { name: "A", in_deck: true },
      ],
    });
    expect(deck.map((c) => c.id)).toEqual(["p0", "p2"]);
    expect(side.map((c) => c.name)).toEqual(["B"]);
    expect(splitPicks({})).toEqual({ deck: [], side: [] });
  });
});

describe("the routes", () => {
  it("parses the list and one draft", () => {
    expect(parseHash("#/drafts")).toEqual({ name: "myDrafts" });
    expect(parseHash(myDraftHash("abc"))).toEqual({ name: "myDrafts", id: "abc" });
  });

  it("asks /me/drafts and /me/drafts/{id}", async () => {
    const asked: string[] = [];
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      asked.push(String(input));
      const body = String(input).endsWith("/me/drafts")
        ? { drafts: [{ id: "d1", name: "Friday draft" }] }
        : { draft: { id: "d1", name: "Friday draft", picks: [] } };
      return new Response(JSON.stringify(body), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }) as never;
    expect((await fetchMyDrafts())[0].id).toBe("d1");
    expect((await fetchMyDraft("d 1")).name).toBe("Friday draft");
    expect(asked).toEqual(["/me/drafts", "/me/drafts/d%201"]);
  });
});
