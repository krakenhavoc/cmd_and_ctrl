import { describe, expect, it } from "vitest";

import { abilityBlocked, exilePermanentShortfall } from "./contextMenu.logic";
import { manaAbilityNeedsPrompt, manaExilePermanentPayment } from "./manaAbilityCost";
import { manaAbilityRider } from "./manaSource";
import type { ManaAbilityView } from "./protocol";

// #1600: "Exile a creature you control" as a cost — The Soul Stone's
// harness and Food Chain's mana ability. The server stamps
// `exile_permanent_label` / `exile_permanent_options` on both ability
// views and reads the answer back as `exile_permanent_ids`.

function manaAbility(extra: Partial<ManaAbilityView>): ManaAbilityView {
  return { index: 0, label: "Add {G}{G}", produced: "{G}{G}", ...extra };
}

describe("exilePermanentShortfall", () => {
  it("is empty when the board can pay", () => {
    expect(exilePermanentShortfall({ cards: ["bear"], min: 1, max: 1 })).toBe("");
  });

  it("names the clause when nothing can pay", () => {
    expect(exilePermanentShortfall({ cards: [], min: 1, max: 1 }, "a creature you control")).toBe(
      "nothing to exile (a creature you control)",
    );
  });

  it("is empty for an ability without the component", () => {
    expect(exilePermanentShortfall(undefined)).toBe("");
  });
});

describe("abilityBlocked with an exile-a-permanent cost", () => {
  it("greys the row with no creature to exile", () => {
    expect(
      abilityBlocked(
        {
          exile_permanent_label: "a creature you control",
          exile_permanent_options: { cards: [], min: 1, max: 1 },
        },
        false,
        false,
      ),
    ).toBe("nothing to exile (a creature you control)");
  });

  it("leaves it live with one", () => {
    expect(
      abilityBlocked(
        { exile_permanent_options: { cards: ["bear"], min: 1, max: 1 } },
        false,
        false,
      ),
    ).toBe("");
  });
});

describe("Food Chain's mana ability", () => {
  const chain = manaAbility({
    exile_permanent_label: "a creature you control",
    exile_permanent_options: { cards: ["bear", "elf"], min: 1, max: 1 },
  });

  it("asks which creature to exile before it is sent", () => {
    expect(manaAbilityNeedsPrompt(chain)).toBe(true);
  });

  it("sends the picks as exile_permanent_ids, never exile_ids", () => {
    expect(manaExilePermanentPayment(["bear"])).toEqual({ exile_permanent_ids: ["bear"] });
    expect(manaExilePermanentPayment([])).toEqual({});
  });

  it("names the cost in the picker's rider", () => {
    expect(manaAbilityRider(chain)).toContain("exile a creature you control");
  });
});
