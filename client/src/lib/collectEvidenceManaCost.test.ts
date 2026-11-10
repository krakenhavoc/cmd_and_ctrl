import { describe, expect, it } from "vitest";
import { manaAbilityNeedsPrompt } from "./manaAbilityCost";
import { manaAbilityRider } from "./manaSource";
import type { ManaAbilityView } from "./protocol";

// Collect evidence N as a mana ability's cost — Cryptex's "{T}, Collect
// evidence 3: Add one mana of any color. Put an unlock counter on this
// artifact." The row ships the any-number exile fields an activated row
// does (Forensic Researcher); the picks ride activate_mana_ability as
// `exile_ids`.

const cryptex: ManaAbilityView = {
  index: 0,
  ref: "own:0",
  label: "{T}, Collect evidence 3: Add one mana of any color.",
  tap_cost: true,
  exile_cost_any_number: true,
  exile_cost_mana_value_min: 3,
  exile_cost_label: "cards from your graveyard with total mana value 3 or greater",
  exile_cost_options: ["g1", "g2"],
  exile_cost_zone: "graveyard",
} as ManaAbilityView;

describe("collect evidence on a mana ability", () => {
  it("needs a prompt even with exactly the cards it could take", () => {
    expect(manaAbilityNeedsPrompt(cryptex)).toBe(true);
  });

  it("names the collection in the source picker's rider", () => {
    expect(manaAbilityRider(cryptex)).toBe("collect evidence 3");
  });
});
