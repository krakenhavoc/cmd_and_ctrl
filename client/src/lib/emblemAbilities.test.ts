import { describe, expect, it } from "vitest";
import { emblemAbilityBlocked, emblemAbilityCost, emblemAsCard } from "./emblemAbilities";
import type { ActivatedAbilityView, EmblemView } from "./protocol";

const row: ActivatedAbilityView = {
  index: 0,
  label: "Tap an untapped artifact you control: This emblem deals 1 damage to any target.",
};

const emblem: EmblemView = {
  instance_id: "e1",
  label: "Karn, Living Legacy emblem",
  text: "Tap an untapped artifact you control: This emblem deals 1 damage to any target.",
  zone_abilities: [row],
};

describe("emblem abilities (#5868)", () => {
  it("stands the emblem in for a card, owned and controlled by its owner", () => {
    const c = emblemAsCard(emblem, "p1");
    expect(c.instance_id).toBe("e1");
    expect(c.name).toBe("Karn, Living Legacy emblem");
    expect(c.owner).toBe("p1");
    expect(c.controller).toBe("p1");
    expect(c.zone_abilities).toEqual([row]);
  });

  it("names the button by the cost", () => {
    expect(emblemAbilityCost(row)).toBe("Tap an untapped artifact you control");
  });

  it("is blocked without priority or on a greyed row", () => {
    expect(emblemAbilityBlocked(row, true)).toBe("");
    expect(emblemAbilityBlocked(row, false)).not.toBe("");
    expect(emblemAbilityBlocked({ ...row, timing_closed: true }, true)).not.toBe("");
  });
});
