// seen.test.ts — `settings.help.seen` (ADR 0125 §3.3, §4): a hint id to
// the version last dismissed; unseen when missing or lower; merged by
// union; retired ids dropped and unknown ids kept.

import { describe, expect, it } from "vitest";

import { isUnseen, normalizeSeen, unionSeen, withSeen, withoutSeen } from "./seen";

describe("the seen map", () => {
  it("a hint is unseen when its id is missing or its stored version is lower", () => {
    expect(isUnseen({}, "table.stack", 1)).toBe(true);
    expect(isUnseen({ "table.stack": 1 }, "table.stack", 1)).toBe(false);
    expect(isUnseen({ "table.stack": 1 }, "table.stack", 2)).toBe(true);
    expect(isUnseen({ "table.stack": 3 }, "table.stack", 2)).toBe(false);
  });

  it("marking seen never lowers a stored version", () => {
    expect(withSeen({}, "a.b", 2)).toEqual({ "a.b": 2 });
    const seen = { "a.b": 3 };
    expect(withSeen(seen, "a.b", 2)).toBe(seen);
  });

  it("forgets the ids it is given and keeps the rest", () => {
    expect(withoutSeen({ "a.b": 1, "a.c": 2, "d.e": 1 }, ["a.b", "a.c"])).toEqual({ "d.e": 1 });
  });

  it("unions two maps, the higher version winning, losing nothing", () => {
    const a = { "lobby.create": 1, "table.stack": 2 };
    const b = { "table.stack": 1, "decks.check": 1 };
    expect(unionSeen(a, b)).toEqual({ "lobby.create": 1, "table.stack": 2, "decks.check": 1 });
    expect(unionSeen(b, a)).toEqual(unionSeen(a, b));
  });

  it("keeps unknown ids, drops retired ones and anything malformed", () => {
    const retired = new Set(["table.old"]);
    const got = normalizeSeen(
      {
        "table.future-thing": 4, // unknown to this client: kept
        "table.old": 1, // retired: dropped
        "lobby.zero": 0,
        "lobby.float": 1.5,
        "lobby.text": "1",
        "": 1,
        "decks.check": 2,
      },
      retired,
    );
    expect(got).toEqual({ "table.future-thing": 4, "decks.check": 2 });
    for (const raw of [null, undefined, 7, "x", [1, 2]]) expect(normalizeSeen(raw)).toEqual({});
  });
});
