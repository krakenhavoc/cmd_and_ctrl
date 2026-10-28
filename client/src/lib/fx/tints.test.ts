// ADR 0209 §4 (#4416): a signature's tint lasts until the turn ends.
// Toxic Deluge's survivors keep a dark-green stain while its -X/-X lasts;
// the client forgets the stain when the turn number or the game changes.

import { get } from "svelte/store";
import { afterEach, describe, expect, it } from "vitest";

import {
  clearTableTints,
  markTints,
  syncTableTints,
  syncTints,
  tableTints,
  tintSurvivors,
  type TintBook,
} from "./tints";

afterEach(() => clearTableTints());

describe("the tint book", () => {
  it("marks cards under a game and turn, and forgets them when the turn moves on", () => {
    let book: TintBook = new Map();
    book = markTints(book, ["a", "b"], "deluge", "g1", 5);
    expect([...book.keys()]).toEqual(["a", "b"]);
    // The same turn keeps them, and keeps the very same book.
    expect(syncTints(book, "g1", 5)).toBe(book);
    expect(syncTints(book, "g1", 6).size).toBe(0);
    // Another game is another table.
    expect(syncTints(book, "g2", 5).size).toBe(0);
  });

  it("keeps a later turn's marks while dropping an earlier turn's", () => {
    let book: TintBook = markTints(new Map(), ["a"], "deluge", "g1", 5);
    book = markTints(book, ["b"], "deluge", "g1", 6);
    expect([...syncTints(book, "g1", 6).keys()]).toEqual(["b"]);
  });
});

describe("the app's tints", () => {
  it("tints the survivors until the turn changes", () => {
    tintSurvivors(["d", "e"], "deluge", "g1", 3);
    expect(get(tableTints)).toEqual(
      new Map([
        ["d", "deluge"],
        ["e", "deluge"],
      ]),
    );
    syncTableTints("g1", 3);
    expect(get(tableTints).size).toBe(2);
    syncTableTints("g1", 4);
    expect(get(tableTints).size).toBe(0);
  });

  it("a frame on the same turn does not wake every tile", () => {
    tintSurvivors(["d"], "deluge", "g1", 3);
    let wakes = 0;
    const stop = tableTints.subscribe(() => wakes++);
    syncTableTints("g1", 3);
    syncTableTints("g1", 3);
    expect(wakes).toBe(1);
    stop();
  });

  it("marking nothing changes nothing", () => {
    tintSurvivors([], "deluge", "g1", 3);
    expect(get(tableTints).size).toBe(0);
  });
});
