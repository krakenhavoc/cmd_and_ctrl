// namedActivator.test.ts — ADR 0106 §1's 2026-10-07 amendment (#1947),
// the client half of "Only your opponents may activate this ability"
// and "Only this creature's owner may activate this ability" as pure
// logic: who sees the row, and who sees it greyed.

import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView, CardView } from "./protocol";
import {
  ROW_NOT_OPEN_TO_YOU,
  abilityRowBlocked,
  mayActivateAcross,
  menuAbilityRows,
  rowOpenToViewer,
} from "./contextMenu.logic";

const ALICE = "alice";
const BOB = "bob";
const CARA = "cara";

const row = (extra: Partial<ActivatedAbilityView>): ActivatedAbilityView => ({
  index: 0,
  ref: "own:0",
  label: "{1}: This creature can't be regenerated this turn.",
  ...extra,
});

const clergy = (controller = ALICE): CardView => ({
  instance_id: "clergy",
  name: "Clergy of the Holy Nimbus",
  owner: ALICE,
  controller,
  type_line: "Creature — Human Cleric",
  activated_abilities: [row({ opponents_only: true })],
});

const incarnation = (controller: string): CardView => ({
  instance_id: "inc",
  name: "Personal Incarnation",
  owner: ALICE,
  controller,
  type_line: "Creature — Avatar Incarnation",
  activated_abilities: [row({ owner_only: true })],
});

const ctx = (card: CardView, viewer: string) => ({
  card,
  tapped: false,
  sick: false,
  loyalty: { card, view: null, viewerID: viewer },
});

describe("opponents-only rows", () => {
  it("are open to every seat but the controller", () => {
    const c = clergy();
    expect(rowOpenToViewer(c.activated_abilities![0], c, ALICE)).toBe(false);
    expect(rowOpenToViewer(c.activated_abilities![0], c, BOB)).toBe(true);
    expect(rowOpenToViewer(c.activated_abilities![0], c, null)).toBe(false);
  });

  it("let an opponent open the permanent, and the controller not", () => {
    expect(mayActivateAcross(clergy(), BOB)).toBe(true);
    expect(mayActivateAcross(clergy(), ALICE)).toBe(false);
    expect(menuAbilityRows(clergy(), BOB).map((a) => a.index)).toEqual([0]);
  });

  it("grey on the controller's own permanent, with a reason", () => {
    const c = clergy();
    expect(abilityRowBlocked(c.activated_abilities![0], "activated", ctx(c, ALICE))).toBe(
      ROW_NOT_OPEN_TO_YOU,
    );
    expect(abilityRowBlocked(c.activated_abilities![0], "activated", ctx(c, BOB))).not.toBe(
      ROW_NOT_OPEN_TO_YOU,
    );
  });
});

describe("owner-only rows", () => {
  it("are the owner's, not the thief's", () => {
    const c = incarnation(BOB);
    expect(rowOpenToViewer(c.activated_abilities![0], c, ALICE)).toBe(true);
    expect(rowOpenToViewer(c.activated_abilities![0], c, BOB)).toBe(false);
    expect(rowOpenToViewer(c.activated_abilities![0], c, CARA)).toBe(false);
  });

  it("let the owner open a stolen permanent, and nobody else", () => {
    expect(mayActivateAcross(incarnation(BOB), ALICE)).toBe(true);
    expect(mayActivateAcross(incarnation(BOB), CARA)).toBe(false);
    expect(menuAbilityRows(incarnation(BOB), CARA)).toEqual([]);
  });

  it("grey for the thief who controls it", () => {
    const c = incarnation(BOB);
    expect(abilityRowBlocked(c.activated_abilities![0], "activated", ctx(c, BOB))).toBe(
      ROW_NOT_OPEN_TO_YOU,
    );
  });

  it("are ordinary on the owner's own permanent", () => {
    const c = incarnation(ALICE);
    expect(abilityRowBlocked(c.activated_abilities![0], "activated", ctx(c, ALICE))).toBe("");
  });
});
