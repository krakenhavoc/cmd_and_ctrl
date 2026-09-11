import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import {
  begin,
  beginForAbility,
  beginForMode,
  cancel,
  canConfirm,
  confirm,
  hasXCost,
  isModal,
  isMultiPick,
  isPicked,
  modeOptionCastable,
  setConfirmHandler,
  togglePick,
  isLegalCardTarget,
  isLegalPlayerTarget,
  legalTargetCount,
  targeting,
} from "./targeting";
import type { ActivatedAbilityView, CardView } from "./protocol";

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "p0", controller: "p0", ...extras };
}

describe("targeting store — S20 legal sets", () => {
  it("uses the server legal set when the card carries one", () => {
    begin(card({ legal_targets: { players: ["p1"], cards: ["bear"] } }), "any");
    const t = get(targeting)!;
    expect(isLegalCardTarget(t, "bear")).toBe(true);
    expect(isLegalCardTarget(t, "rock")).toBe(false);
    expect(isLegalPlayerTarget(t, "p1")).toBe(true);
    expect(isLegalPlayerTarget(t, "p0")).toBe(false);
    expect(legalTargetCount(t)).toBe(2);
    cancel();
    expect(get(targeting)).toBeNull();
  });

  it("falls back to mode heuristics for free-form cards", () => {
    begin(card(), "creature");
    const t = get(targeting)!;
    expect(t.legal).toBeUndefined();
    expect(isLegalCardTarget(t, "anything")).toBe(true);
    expect(isLegalPlayerTarget(t, "p1")).toBe(false);
    expect(legalTargetCount(t)).toBe(-1);
    begin(card(), "player");
    expect(isLegalPlayerTarget(get(targeting)!, "p1")).toBe(true);
    cancel();
  });
});

describe("hasXCost + xValue on the prompt — S20 sub-PR 3", () => {
  it("detects {X} in the printed cost", () => {
    expect(hasXCost(card({ mana_cost: "{X}{R}" }))).toBe(true);
    expect(hasXCost(card({ mana_cost: "{2}{U}" }))).toBe(false);
    expect(hasXCost(card({}))).toBe(false);
  });

  it("carries the announced X through the targeting prompt", () => {
    begin(card({ mana_cost: "{X}{R}", legal_targets: { players: ["p1"] } }), "any", { xValue: 4 });
    expect(get(targeting)?.choices?.xValue).toBe(4);
    cancel();
  });
});

// --- S20 sub-PR 4: modal spells ----------------------------------

describe("modal targeting", () => {
  const charm = {
    instance_id: "c-charm",
    name: "Rakdos Charm",
    modes: {
      prompt: "Choose one",
      min: 1,
      max: 1,
      options: [
        {
          label: "Exile target player's graveyard.",
          target_mode: "player",
          legal_targets: { players: ["p1"] },
        },
        {
          label: "Destroy target artifact.",
          target_mode: "permanent",
          legal_targets: { cards: [] },
        },
        { label: "Each creature deals 1 damage to its controller." },
      ],
    },
  } as unknown as CardView;

  it("isModal / modeOptionCastable", () => {
    expect(isModal(charm)).toBe(true);
    expect(isModal({ instance_id: "x", name: "Bolt" } as unknown as CardView)).toBe(false);
    expect(modeOptionCastable(charm.modes!.options[0])).toBe(true);
    expect(modeOptionCastable(charm.modes!.options[1])).toBe(false);
    expect(modeOptionCastable(charm.modes!.options[2])).toBe(true);
  });

  it("beginForMode takes the option's legal set and label, and carries the modes", () => {
    beginForMode(charm, charm.modes!.options[0], [0], { xValue: 3 });
    const t = get(targeting)!;
    expect(t.mode).toBe("player");
    expect(t.label).toBe("Exile target player's graveyard.");
    expect(t.modes).toEqual([0]);
    expect(t.choices?.xValue).toBe(3);
    expect(isLegalPlayerTarget(t, "p1")).toBe(true);
    expect(isLegalPlayerTarget(t, "p0")).toBe(false);
    expect(isLegalCardTarget(t, "c-anything")).toBe(false);
    cancel();
  });
});

// --- S21 sub-PR 2: activated abilities ----------------------------

describe("activated-ability targeting", () => {
  const bombardment = {
    instance_id: "c-bomb",
    name: "Goblin Bombardment",
  } as unknown as CardView;
  const ability = {
    index: 0,
    label: "Sacrifice a creature: deal 1 damage to any target",
    sacrifice_label: "a creature",
    sacrifice_options: { cards: ["c-fodder"] },
    target_mode: "any",
    legal_targets: { players: ["p1"], cards: ["c-bear"] },
  } as unknown as ActivatedAbilityView;

  it("carries the ability index and the paid sacrifice through the prompt", () => {
    beginForAbility(bombardment, ability, ["c-fodder"]);
    const t = get(targeting)!;
    expect(t.mode).toBe("any");
    // crewIDs defaults to empty: this ability has no crew component,
    // and the server treats absent and empty the same.
    expect(t.ability).toEqual({ index: 0, sacrificeIDs: ["c-fodder"], crewIDs: [] });
    // The legal set comes from the ability, not the source card.
    expect(isLegalPlayerTarget(t, "p1")).toBe(true);
    expect(isLegalCardTarget(t, "c-bear")).toBe(true);
    expect(isLegalCardTarget(t, "c-fodder")).toBe(false);
    cancel();
    expect(get(targeting)).toBeNull();
  });
});

// --- S20 sub-PR 5: multi-target ------------------------------------

describe("multi-pick targeting", () => {
  const pair = {
    instance_id: "c-ashes",
    name: "Ashes to Ashes",
    legal_targets: { cards: ["c-a", "c-b", "c-c"], min: 2, max: 2 },
  } as unknown as CardView;

  it("single-target prompts are not multi-pick", () => {
    begin(
      {
        instance_id: "c-bolt",
        name: "Bolt",
        legal_targets: { players: ["p1"] },
      } as unknown as CardView,
      "any",
    );
    expect(isMultiPick(get(targeting)!)).toBe(false);
    cancel();
  });

  it("toggles picks in click order, caps at max, confirms at min", () => {
    begin(pair, "creature");
    let t = get(targeting)!;
    expect(isMultiPick(t)).toBe(true);
    expect(canConfirm(t)).toBe(false);
    t = togglePick(t, { kind: "card", id: "c-b" });
    t = togglePick(t, { kind: "card", id: "c-a" });
    expect(t.picked.map((p) => p.id)).toEqual(["c-b", "c-a"]);
    expect(canConfirm(t)).toBe(true);
    // Third pick refused at max 2.
    t = togglePick(t, { kind: "card", id: "c-c" });
    expect(t.picked.length).toBe(2);
    // Toggling a picked one removes it.
    t = togglePick(t, { kind: "card", id: "c-b" });
    expect(t.picked.map((p) => p.id)).toEqual(["c-a"]);
    expect(isPicked(t, "c-a")).toBe(true);
    expect(isPicked(t, "c-b")).toBe(false);
    expect(canConfirm(t)).toBe(false);
    cancel();
  });

  it("up to N confirms with nothing picked; unbounded max never caps", () => {
    begin(
      { ...pair, legal_targets: { cards: ["c-a"], min: 0, max: 2 } } as unknown as CardView,
      "permanent",
    );
    expect(canConfirm(get(targeting)!)).toBe(true);
    cancel();
    begin(
      {
        ...pair,
        legal_targets: { cards: ["c-a", "c-b", "c-c"], min: 1, max: 0 },
      } as unknown as CardView,
      "creature",
    );
    let t = get(targeting)!;
    for (const id of ["c-a", "c-b", "c-c"]) t = togglePick(t, { kind: "card", id });
    expect(t.picked.length).toBe(3);
    cancel();
  });

  it("confirm() routes to the registered handler", () => {
    let fired = 0;
    setConfirmHandler(() => fired++);
    confirm();
    expect(fired).toBe(1);
    setConfirmHandler(null);
    confirm();
    expect(fired).toBe(1);
  });
});
