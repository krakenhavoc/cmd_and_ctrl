// ADR 0114 owner decision 1 (#2076): what the Ring chip says. The
// render test (ringDisplay.render.test.ts) checks the markup; this
// pins the words.

import { describe, it, expect } from "vitest";

import {
  emblemChipName,
  emblemDescription,
  emblemLines,
  isLevelledEmblem,
  ordinal,
  pendingNote,
  ringBearerNames,
  ringBearerTitle,
} from "./ringEmblem";
import type { CardView, EmblemView, PlayerView } from "./protocol";

const LINES = [
  { text: "Whenever your Ring-bearer attacks, draw a card, then discard a card.", at: 2 },
  {
    text: "Your Ring-bearer is legendary and can't be blocked by creatures with greater power.",
    at: 1,
  },
  {
    text: "Whenever your Ring-bearer deals combat damage to a player, each opponent loses 3 life.",
    at: 4,
  },
  {
    text: "Whenever your Ring-bearer becomes blocked by a creature, the blocking creature's controller sacrifices it at end of combat.",
    at: 3,
  },
];

const ring = (level: number): EmblemView => ({
  instance_id: "e1",
  label: "The Ring",
  text: "",
  level,
  lines: LINES,
});

describe("the Ring's lines", () => {
  it("lists every line in the order it is gained, gained up to the level", () => {
    const lines = emblemLines(ring(2));
    expect(lines.map((l) => l.at)).toEqual([1, 2, 3, 4]);
    expect(lines.map((l) => l.gained)).toEqual([true, true, false, false]);
  });

  it("has every line at four temptations, and more", () => {
    expect(emblemLines(ring(4)).every((l) => l.gained)).toBe(true);
    expect(emblemLines(ring(6)).every((l) => l.gained)).toBe(true);
  });

  it("falls back to the text as one gained line with no lines on the wire", () => {
    expect(emblemLines({ text: "Old text.", level: 1 })).toEqual([
      { text: "Old text.", at: 1, gained: true },
    ]);
  });

  it("is levelled only when the server sends a level or lines", () => {
    expect(isLevelledEmblem(ring(1))).toBe(true);
    expect(isLevelledEmblem({ level: 2 })).toBe(true);
    expect(isLevelledEmblem({})).toBe(false);
    expect(isLevelledEmblem({ lines: [] })).toBe(false);
  });
});

describe("the words", () => {
  it("says ordinals", () => {
    expect([1, 2, 3, 4, 11, 12, 13, 21, 22, 23].map(ordinal)).toEqual([
      "1st",
      "2nd",
      "3rd",
      "4th",
      "11th",
      "12th",
      "13th",
      "21st",
      "22nd",
      "23rd",
    ]);
  });

  it("marks a line not yet gained by when it will be", () => {
    expect(pendingNote(4)).toBe("after the 4th temptation");
  });

  it("names the chip by its level", () => {
    expect(emblemChipName(ring(1))).toBe("The Ring, tempted once");
    expect(emblemChipName(ring(3))).toBe("The Ring, tempted 3 times");
  });

  it("describes every line and whether it is gained", () => {
    const d = emblemDescription(ring(3));
    expect(d).toContain("Gained: Your Ring-bearer is legendary");
    expect(d).toContain("Gained: Whenever your Ring-bearer attacks");
    expect(d).toContain("Gained: Whenever your Ring-bearer becomes blocked");
    expect(d).toContain("Not yet, after the 4th temptation: Whenever your Ring-bearer deals");
  });

  it("titles the marker with its controller", () => {
    expect(ringBearerTitle("Alice")).toBe("Alice's Ring-bearer");
    expect(ringBearerTitle(undefined)).toBe("Ring-bearer");
  });
});

describe("ringBearerNames", () => {
  it("names the controller of each Ring-bearer and nothing else", () => {
    const cards = [
      { instance_id: "a", controller: "p1", ring_bearer: true },
      { instance_id: "b", controller: "p1" },
      { instance_id: "c", controller: "p2", ring_bearer: true },
    ] as CardView[];
    const seats = [
      { id: "p1", name: "Alice" },
      { id: "p2", name: "Bob" },
    ] as PlayerView[];
    expect(ringBearerNames(cards, seats)).toEqual({ a: "Alice", c: "Bob" });
  });
});
