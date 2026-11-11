// @vitest-environment jsdom
// #5699: the picks split into Main and Sideboard, kept in this browser.

import { beforeEach, describe, expect, it, vi } from "vitest";

import type { DraftCard } from "./draft";
import {
  forgetSplit,
  loadSplit,
  moveTo,
  pruneSide,
  pruneSplits,
  saveSplit,
  splitCards,
} from "./draftSplit";

const card = (id: string): DraftCard => ({ id, name: id });
const cards = ["a", "b", "c", "d"].map(card);
const ids = (cs: DraftCard[]): string[] => cs.map((c) => c.id);

describe("the split of the picks", () => {
  it("puts every card in Main when nothing is set aside", () => {
    const s = splitCards(cards, []);
    expect(ids(s.main)).toEqual(["a", "b", "c", "d"]);
    expect(s.side).toEqual([]);
  });

  it("puts a card that has just been picked in Main", () => {
    const s = splitCards([...cards, card("new")], ["b"]);
    expect(ids(s.side)).toEqual(["b"]);
    expect(ids(s.main)).toEqual(["a", "c", "d", "new"]);
  });

  it("ignores a sideboard id that is not a card, and keeps the order of the cards", () => {
    const s = splitCards(cards, ["gone", "d", "a"]);
    expect(ids(s.side)).toEqual(["a", "d"]);
    expect(ids(s.main)).toEqual(["b", "c"]);
  });

  it("moves a card to the Sideboard and back, each once", () => {
    let side: string[] = [];
    side = moveTo(side, "b", "side");
    side = moveTo(side, "b", "side");
    expect(side).toEqual(["b"]);
    side = moveTo(side, "c", "side");
    expect(side).toEqual(["b", "c"]);
    side = moveTo(side, "b", "main");
    expect(side).toEqual(["c"]);
    expect(moveTo(side, "zz", "main")).toBe(side);
  });

  it("drops ids that are not cards and repeats", () => {
    expect(pruneSide(cards, ["a", "x", "a", "d"])).toEqual(["a", "d"]);
  });
});

describe("keeping the split", () => {
  beforeEach(() => localStorage.clear());

  it("reads back what it saved, per draft", () => {
    saveSplit("d1", ["a", "b"]);
    saveSplit("d2", ["z"]);
    expect(loadSplit("d1")).toEqual(["a", "b"]);
    expect(loadSplit("d2")).toEqual(["z"]);
    expect(loadSplit("d3")).toEqual([]);
  });

  it("stores nothing for an empty sideboard, and clears what was there", () => {
    saveSplit("d1", ["a"]);
    saveSplit("d1", []);
    expect(localStorage.getItem("cmdctrl.draftSplit.d1")).toBeNull();
    expect(loadSplit("d1")).toEqual([]);
  });

  it("stores only card ids and a time, nothing about a person", () => {
    saveSplit("d1", ["a"], 1000);
    expect(JSON.parse(localStorage.getItem("cmdctrl.draftSplit.d1")!)).toEqual({
      side: ["a"],
      at: 1000,
    });
  });

  it("reads bad data as an empty split", () => {
    localStorage.setItem("cmdctrl.draftSplit.bad", "{not json");
    localStorage.setItem("cmdctrl.draftSplit.shape", JSON.stringify({ side: "a" }));
    localStorage.setItem("cmdctrl.draftSplit.mixed", JSON.stringify({ side: ["a", 3, null] }));
    expect(loadSplit("bad")).toEqual([]);
    expect(loadSplit("shape")).toEqual([]);
    expect(loadSplit("mixed")).toEqual(["a"]);
  });

  it("forgets a stale split on reading it", () => {
    saveSplit("d1", ["a"], 1000);
    expect(loadSplit("d1", 1000 + 15 * 24 * 3600 * 1000)).toEqual([]);
    expect(localStorage.getItem("cmdctrl.draftSplit.d1")).toBeNull();
  });

  it("forgets one draft's split", () => {
    saveSplit("d1", ["a"]);
    saveSplit("d2", ["b"]);
    forgetSplit("d1");
    expect(loadSplit("d1")).toEqual([]);
    expect(loadSplit("d2")).toEqual(["b"]);
  });

  it("prunes the splits of drafts this browser has no seat in, and nothing else", () => {
    saveSplit("live", ["a"]);
    saveSplit("over", ["b"]);
    localStorage.setItem("cmdctrl.draft.live", JSON.stringify({ token: "t" }));
    localStorage.setItem("other", "x");
    pruneSplits(["live"]);
    expect(loadSplit("live")).toEqual(["a"]);
    expect(loadSplit("over")).toEqual([]);
    expect(localStorage.getItem("cmdctrl.draft.live")).not.toBeNull();
    expect(localStorage.getItem("other")).toBe("x");
  });

  it("prunes a split that has gone stale even when its draft is live", () => {
    saveSplit("live", ["a"], 1000);
    pruneSplits(["live"], 1000 + 15 * 24 * 3600 * 1000);
    expect(localStorage.getItem("cmdctrl.draftSplit.live")).toBeNull();
  });

  it("works when storage throws", () => {
    const boom = vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("denied");
    });
    const set = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("denied");
    });
    expect(loadSplit("d1")).toEqual([]);
    expect(() => saveSplit("d1", ["a"])).not.toThrow();
    expect(() => forgetSplit("d1")).not.toThrow();
    expect(() => pruneSplits([])).not.toThrow();
    boom.mockRestore();
    set.mockRestore();
  });
});
