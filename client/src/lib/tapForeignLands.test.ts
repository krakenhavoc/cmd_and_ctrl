// tapForeignLands.test.ts — #6204 (ADR 0106 §1 amendment 2026-11-25),
// the client half of "Until end of turn, you may tap lands you don't
// control for mana" (Piracy). The server marks the {T} row of another
// player's land `any_player` on the holder's copy only, with the
// "purpose:cast" tag its mana will carry; the existing any-player path
// then lets that viewer tap the land, and nobody else.

import { describe, expect, it } from "vitest";
import type { CardView, ManaAbilityView } from "./protocol";
import { mayActivateAcross, menuManaRows } from "./contextMenu.logic";

const ME = "me";
const ALICE = "alice";

const tapRow = (extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index: 0,
  ref: "own:0",
  label: "{T}: Add {R}.",
  produced: "{R}",
  tap_cost: true,
  ...extra,
});

const mountain = (row: Partial<ManaAbilityView>): CardView => ({
  instance_id: "mountain",
  name: "Mountain",
  owner: ALICE,
  controller: ALICE,
  type_line: "Basic Land — Mountain",
  mana_abilities: [tapRow(row)],
});

describe("tapping a land you don't control (#6204)", () => {
  it("the holder's copy opens the land's {T} row to them", () => {
    const card = mountain({ any_player: true, restrictions: ["purpose:cast"] });
    expect(mayActivateAcross(card, ME)).toBe(true);
    expect(menuManaRows(card, ME).map((r) => r.ref)).toEqual(["own:0"]);
  });

  it("everyone else's copy is the public row, which they can't tap", () => {
    const card = mountain({});
    expect(mayActivateAcross(card, ME)).toBe(false);
    expect(menuManaRows(card, ME)).toEqual([]);
  });

  it("the controller keeps the whole menu", () => {
    const card = mountain({});
    expect(menuManaRows(card, ALICE).map((r) => r.ref)).toEqual(["own:0"]);
  });
});
