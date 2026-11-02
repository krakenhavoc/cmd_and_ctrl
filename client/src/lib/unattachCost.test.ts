import { describe, expect, it } from "vitest";
import { get } from "svelte/store";

import { abilityBlocked, unattachShortfall } from "./contextMenu.logic";
import type { ActivatedAbilityView, CardView } from "./protocol";
import { beginForAbility, divideTotal, targeting } from "./targeting";

// #4799: "Unattach an Equipment from <this>" as a cost (Captain America,
// First Avenger). The server stamps `unattach_label` / `unattach_options`
// on the row and, on a divided clause whose amount is read from the pick,
// `divide.by_unattach`; the client asks for the Equipment first and sizes
// the split from it.

const throwRow: ActivatedAbilityView = {
  index: 0,
  label: "{3}, Unattach an Equipment from Captain America: He deals damage …",
  unattach_label: "an Equipment from Captain America, First Avenger",
  unattach_options: { cards: ["knife", "hammer"], min: 1, max: 1 },
  legal_targets: {
    players: ["opp"],
    cards: ["a", "b", "c"],
    min: 1,
    max: 3,
    divide: { by_unattach: { knife: 1, hammer: 4 } },
  },
};

function cap(): CardView {
  return {
    instance_id: "cap",
    name: "Captain America, First Avenger",
    type_line: "Legendary Creature — Human Soldier Hero",
    owner: "me",
    controller: "me",
  } as CardView;
}

describe("divideTotal with by_unattach", () => {
  it("is the amount quoted for the picked Equipment", () => {
    const d = { by_unattach: { knife: 1, hammer: 4 } };
    expect(divideTotal(d, undefined, ["hammer"])).toBe(4);
    expect(divideTotal(d, undefined, ["knife"])).toBe(1);
  });
  it("is 0 with no pick, whatever the fixed total says", () => {
    expect(divideTotal({ total: 7, by_unattach: { knife: 1 } }, undefined)).toBe(0);
  });
});

describe("beginForAbility sizes the division from the unattach pick", () => {
  it("divides the Hammer's 4", () => {
    beginForAbility(cap(), throwRow, [], [], undefined, undefined, undefined, undefined, [
      "hammer",
    ]);
    expect(get(targeting)?.steps[0]?.divide).toBe(4);
    targeting.set(null);
  });
  it("divides the Knife's 1", () => {
    beginForAbility(cap(), throwRow, [], [], undefined, undefined, undefined, undefined, ["knife"]);
    expect(get(targeting)?.steps[0]?.divide).toBe(1);
    targeting.set(null);
  });
});

describe("an unattach cost with nothing attached", () => {
  it("greys the row", () => {
    expect(
      unattachShortfall({ cards: [], min: 1, max: 1 }, "an Equipment from Captain America"),
    ).toBe("nothing to unattach (an Equipment from Captain America)");
    expect(
      abilityBlocked(
        { unattach_label: "an Equipment from Cap", unattach_options: { min: 1, max: 1 } },
        false,
        false,
      ),
    ).toBe("nothing to unattach (an Equipment from Cap)");
  });
  it("leaves it live with one attached, and without the component", () => {
    expect(unattachShortfall({ cards: ["knife"], min: 1, max: 1 })).toBe("");
    expect(unattachShortfall(undefined)).toBe("");
  });
});
