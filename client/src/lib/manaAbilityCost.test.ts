import { describe, expect, it } from "vitest";

import {
  chosenPlayerOptions,
  manaAbilityNeedsPrompt,
  manaChosenPlayerPayment,
  manaColorsForChosenPlayer,
  manaTapNeedsPicker,
  manaTapPayment,
  manaXPayment,
} from "./manaAbilityCost";
import { modesUnderChoices } from "./targeting";
import type { CardView } from "./protocol";
import type { ManaAbilityView } from "./protocol";

function ability(extra: Partial<ManaAbilityView>): ManaAbilityView {
  return { index: 0, label: "Add {C}", produced: "{C}", ...extra };
}

describe("manaAbilityNeedsPrompt", () => {
  it("sends a plain tap ability straight out", () => {
    expect(manaAbilityNeedsPrompt(ability({}))).toBe(false);
  });

  // #1283: Cadaverous Bloom asks which card leaves the hand.
  it("asks before an exile-a-card cost", () => {
    expect(
      manaAbilityNeedsPrompt(ability({ exile_cost_n: 1, exile_cost_options: ["a", "b"] })),
    ).toBe(true);
  });

  // #1213: the panel click path used to skip this and send no
  // discard_ids, which the server refuses.
  it("asks before a discard cost", () => {
    expect(
      manaAbilityNeedsPrompt(ability({ discard_cost_n: 1, discard_cost_options: ["a"] })),
    ).toBe(true);
  });

  // #758: Springleaf Drum asks which untapped creature pays the
  // second half of its activation cost.
  it("asks before a tap-another cost", () => {
    expect(
      manaAbilityNeedsPrompt(ability({ tap_others_options: { cards: ["bear"], min: 1, max: 1 } })),
    ).toBe(true);
  });

  // #6085: Rosheen, Roaring Prophet asks which cards with {X} to reveal
  // while the hand holds one, and taps straight out (revealing none) when
  // it holds none.
  it("asks before Rosheen's reveal only when there is a card to show", () => {
    expect(
      manaAbilityNeedsPrompt(
        ability({ reveal_cost_any_number: true, reveal_cost_options: ["fireball"] }),
      ),
    ).toBe(true);
    expect(manaAbilityNeedsPrompt(ability({ reveal_cost_any_number: true }))).toBe(false);
  });
});

describe("manaTapPayment", () => {
  it("sends the chosen permanents as tap_ids", () => {
    expect(manaTapPayment(["bear", "elf"])).toEqual({ tap_ids: ["bear", "elf"] });
  });

  it("keeps ordinary mana-ability payloads unchanged", () => {
    expect(manaTapPayment([])).toEqual({});
  });
});

// #5447: Hazel of the Rootbloom's "Tap X untapped tokens you control".
describe("manaTapNeedsPicker", () => {
  it("sends a fixed count on its own when the options are exactly that many", () => {
    expect(manaTapNeedsPicker({ cards: ["bear"], min: 1, max: 1 })).toBe(false);
    expect(manaTapNeedsPicker({ cards: ["a", "b"], min: 2, max: 2 })).toBe(false);
  });

  it("asks which permanents pay a fixed count with more options than it needs", () => {
    expect(manaTapNeedsPicker({ cards: ["bear", "elf"], min: 1, max: 1 })).toBe(true);
  });

  it("always asks how many to tap when the count is the player's", () => {
    expect(manaTapNeedsPicker({ cards: ["a"], min: 0, max: 1 })).toBe(true);
    expect(manaTapNeedsPicker({ cards: ["a", "b", "c"], min: 0, max: 3 })).toBe(true);
  });

  it("asks nothing with nothing to tap or no such cost", () => {
    expect(manaTapNeedsPicker({ cards: [], min: 0, max: 0 })).toBe(false);
    expect(manaTapNeedsPicker(undefined)).toBe(false);
  });
});

// #6125: X in a mana ability's cost.
describe("manaXPayment", () => {
  it("asks before an {X} row and sends the X named", () => {
    const rockets = ability({ demands_x: true, mana_cost: "{X}" });
    expect(manaAbilityNeedsPrompt(rockets)).toBe(true);
    expect(manaXPayment(rockets, 3, undefined)).toEqual({ x_value: 3 });
    expect(manaXPayment(rockets, 0, undefined)).toEqual({});
  });

  it("takes a sacrifice-X row's X from the permanents sacrificed", () => {
    const pasture = ability({
      demands_x: true,
      sacrifice_options: { cards: ["g1", "g2"], count_from_x: true },
    });
    expect(manaXPayment(pasture, undefined, ["g1", "g2"])).toEqual({ x_value: 2 });
  });

  it("sends nothing for a row with no X", () => {
    expect(manaXPayment(ability({}), 2, undefined)).toEqual({});
  });
});

// #6125: "Choose X" (Doomsday Confluence) takes exactly the X announced.
describe("modesUnderChoices for Choose X", () => {
  it("bounds the picker at the announced X", () => {
    const card = {
      instance_id: "c",
      name: "Doomsday Confluence",
      modes: {
        prompt: "Choose X",
        min: 0,
        max: 0,
        repeatable: true,
        count_is_x: true,
        options: [],
      },
    } as unknown as CardView;
    const m = modesUnderChoices(card, { xValue: 3 }).modes;
    expect(m?.min).toBe(3);
    expect(m?.max).toBe(3);
  });
});

// #6101: Spectral Searchlight's "Choose a player".
describe("a chosen-player mana row", () => {
  it("asks before it is sent", () => {
    expect(manaAbilityNeedsPrompt(ability({ chosen_player: true }))).toBe(true);
  });

  it("offers every player still in the game, you first", () => {
    const seats = [
      { id: "a", name: "Ann" },
      { id: "b", name: "Bea" },
      { id: "c", name: "Cy", eliminated: true },
      { id: "d", name: "Dee" },
    ];
    expect(chosenPlayerOptions(seats, "b")).toEqual([
      { id: "b", name: "Bea", you: true },
      { id: "a", name: "Ann", you: false },
      { id: "d", name: "Dee", you: false },
    ]);
  });

  it("sends chosen_player, and a colour only for yourself", () => {
    expect(manaChosenPlayerPayment(undefined)).toEqual({});
    expect(manaChosenPlayerPayment("b")).toEqual({ chosen_player: "b" });
    expect(manaColorsForChosenPlayer(["G"], "a", "a")).toEqual(["G"]);
    expect(manaColorsForChosenPlayer(["G"], undefined, "a")).toEqual(["G"]);
    expect(manaColorsForChosenPlayer(["G"], "b", "a")).toEqual([]);
  });
});
