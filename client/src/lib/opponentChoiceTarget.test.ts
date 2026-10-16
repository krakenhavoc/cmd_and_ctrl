import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import { begin, beginForAbility, cancel, stepsFor, targeting } from "./targeting";
import type { ActivatedAbilityView, CardView, LegalTargetsView } from "./protocol";

// #2055: a clause "of an opponent's choice" is not the controller's to
// pick. The walk skips it — the other steps keep their own slots — and
// a statement made only of such clauses has no walk at all, so the
// announcement is sent at once and the server asks the opponent.

const mine: LegalTargetsView = {
  cards: ["c1"],
  min: 1,
  max: 1,
  label: "target creature you control",
};
const theirs: LegalTargetsView = {
  cards: ["c2"],
  min: 1,
  max: 1,
  label: "target creature of an opponent's choice they control",
  chosen_by_opponent: true,
};

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "src", name: "Arena", owner: "p0", controller: "p0", ...extras };
}

describe("stepsFor — an opponent's clause is not a step", () => {
  it("skips the opponent's clause and keeps the controller's slot numbers", () => {
    const steps = stepsFor("creature", undefined, [theirs, mine], 0);
    expect(steps).toHaveLength(1);
    expect(steps[0].slot).toBe(1);
    expect(steps[0].label).toBe("target creature you control");
  });

  it("is an empty walk when every clause is an opponent's choice", () => {
    expect(stepsFor("creature", theirs, undefined, 0)).toEqual([]);
  });
});

describe("begin / beginForAbility — nothing to pick", () => {
  it("reports that a spell with only an opponent's clause opens no walk", () => {
    cancel();
    const evangelize = card({ name: "Evangelize", target_mode: "creature", legal_targets: theirs });
    expect(begin(evangelize, "creature")).toBe(false);
    expect(get(targeting)).toBeNull();
  });

  it("opens the walk on the controller's own clause only", () => {
    cancel();
    const ability = {
      index: 0,
      label:
        "{3}, {T}: Tap target creature you control and target creature of an opponent's choice they control.",
      target_mode: "creature",
      clauses: [mine, theirs],
    } as unknown as ActivatedAbilityView;
    beginForAbility(card(), ability, []);
    const t = get(targeting);
    expect(t?.steps).toHaveLength(1);
    expect(t?.steps[0].slot).toBe(0);
    cancel();
  });

  it("leaves no walk open for an ability whose only clause is an opponent's", () => {
    cancel();
    const ability = {
      index: 0,
      label: "{3}, {T}: … target creature of an opponent's choice instead.",
      target_mode: "creature",
      legal_targets: theirs,
    } as unknown as ActivatedAbilityView;
    beginForAbility(card({ name: "Nova Pentacle" }), ability, []);
    expect(get(targeting)).toBeNull();
  });
});
