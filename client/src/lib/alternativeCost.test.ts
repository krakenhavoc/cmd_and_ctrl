import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import {
  altCostPayCount,
  altCostPayOptions,
  alternativeCostByKey,
  alternativeCostsOf,
  applyCastChoices,
  begin,
  cancel,
  targeting,
} from "./targeting";
import { canCastFromHand } from "./timing";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "p0", controller: "p0", ...extras };
}

// Cyclonic Rift, the shape that motivated S22: a printed clause that
// targets one opposing permanent, and an overload cost that deletes
// the clause entirely.
const rift = card({
  instance_id: "rift",
  name: "Cyclonic Rift",
  type_line: "Instant",
  mana_cost: "{1}{U}",
  target_mode: "permanent",
  legal_targets: { cards: [], min: 1, max: 1 },
  alternative_costs: [{ key: "overload", label: "Overload {6}{U}", mana_cost: "{6}{U}" }],
});

// Wash Away, the other shape: the cost swaps in a WIDER clause
// rather than removing it, so cleaved the picker offers more.
const washAway = card({
  instance_id: "wash",
  name: "Wash Away",
  type_line: "Instant",
  mana_cost: "{U}",
  target_mode: "stack_spell",
  legal_targets: { cards: [], min: 1, max: 1 },
  alternative_costs: [
    {
      key: "cleave",
      label: "Cleave {1}{U}{U}",
      mana_cost: "{1}{U}{U}",
      target_mode: "stack_spell",
      legal_targets: { cards: ["s-bolt"], min: 1, max: 1 },
    },
  ],
});

describe("alternative costs on cast — S22", () => {
  it("reads the offers off the wire, and defaults to none", () => {
    expect(alternativeCostsOf(card())).toEqual([]);
    expect(alternativeCostsOf(rift).map((a) => a.key)).toEqual(["overload"]);
  });

  it("looks an offer up by key; undefined means the printed mana cost", () => {
    expect(alternativeCostByKey(rift, "overload")?.mana_cost).toBe("{6}{U}");
    expect(alternativeCostByKey(rift, undefined)).toBeUndefined();
    expect(alternativeCostByKey(rift, "evoke")).toBeUndefined();
  });

  it("takes the alternative cost's legal set, not the card's", () => {
    // Hard-cast, the printed clause has nothing legal on the stack.
    begin(washAway, "stack_spell", {});
    expect(get(targeting)!.legal!.cards.size).toBe(0);
    cancel();
    // Cleaved, the wider clause can hit the Bolt.
    begin(washAway, "stack_spell", { altCost: "cleave" }, alternativeCostByKey(washAway, "cleave"));
    const t = get(targeting)!;
    expect(t.legal!.cards.has("s-bolt")).toBe(true);
    expect(t.choices?.altCost).toBe("cleave");
    cancel();
  });

  it("puts the chosen cost on the cast payload, and omits it otherwise", () => {
    const paid: Record<string, unknown> = {};
    applyCastChoices(paid, { altCost: "overload", xValue: 2 });
    expect(paid).toEqual({ alternative_cost: "overload", x_value: 2 });

    const plain: Record<string, unknown> = {};
    applyCastChoices(plain, {});
    expect(plain).toEqual({});
  });
});

// An alternative cost that rewrites the target clause also rewrites
// castability: a Rift with nothing to bounce is still overloadable.
describe("canCastFromHand — alternative costs", () => {
  function zone(kind: string, cards: CardView[] = []): ZoneView {
    return { kind, owner: "p0", count: cards.length, cards };
  }

  // S31: the cast verdict is the server's enumerated move list, so
  // the fixture states what the server offered. `moves` is the
  // instance IDs the enumerator listed a cast for.
  function snapshot(hand: CardView[], moves: string[] = []): GameView {
    const seat: PlayerView = {
      id: "p0",
      name: "Me",
      seat: 0,
      life: 40,
      library: zone("library"),
      hand: zone("hand", hand),
      graveyard: zone("graveyard"),
      command: zone("command"),
      commander_damage: {},
      life_history: [],
    };
    return {
      id: "g",
      state: "active",
      seats: [seat],
      battlefield: zone("battlefield"),
      stack: zone("stack"),
      exile: zone("exile"),
      turn: {
        number: 1,
        active_seat: 0,
        priority_holder: 0,
        phase: "precombat_main",
        step: "precombat_main",
      },
      mulligans_open: false,
      stack_items: [],
      split_second_active: false,
      legal_moves: [
        { type: "pass_priority", player: "p0", kind: "pass", label: "Pass priority" },
        ...moves.map((id) => ({
          type: "cast_spell",
          player: "p0",
          kind: "cast" as const,
          label: `Cast ${id}`,
          source: id,
        })),
      ],
    };
  }

  it("stays castable with an empty printed legal set when a cost clears the clause", () => {
    expect(canCastFromHand(rift, snapshot([rift], ["rift"]), "p0").legal).toBe(true);
  });

  it("still denies a card whose every cost option has nothing to point at", () => {
    const verdict = canCastFromHand(washAway, snapshot([washAway], ["wash"]), "p0");
    // The cleave clause has a legal spell, so the server offered it —
    // flip the cleave set empty, withhold the move, and the denial
    // comes back with the reason the alternative-cost branch supplies.
    expect(verdict.legal).toBe(true);
    const stranded = {
      ...washAway,
      alternative_costs: [
        { key: "cleave", label: "Cleave {1}{U}{U}", legal_targets: { cards: [], min: 1, max: 1 } },
      ],
    };
    const denied = canCastFromHand(stranded, snapshot([stranded]), "p0");
    expect(denied.legal).toBe(false);
    expect(denied.reason).toBe("No legal target");
  });

  it("leaves a card with no alternative costs on the plain target branch", () => {
    const bolt = card({
      instance_id: "bolt",
      type_line: "Instant",
      legal_targets: { cards: [], min: 1, max: 1 },
    });
    expect(canCastFromHand(bolt, snapshot([bolt]), "p0").reason).toBe("No legal target");
  });
});

// S28: an alternative cost can charge a CARD as well as (or instead
// of) mana. The client has to tell three cases apart, and the
// difference between the last two is the whole point of the helper:
// "charges no cards" must not look like "charges a card you cannot
// pay", because one skips the picker and the other opens it on an
// empty list.
describe("altCostPayOptions", () => {
  const forceOfWill = card({
    instance_id: "fow",
    name: "Force of Will",
    type_line: "Instant",
    mana_cost: "{3}{U}{U}",
    target_mode: "stack_spell",
    alternative_costs: [
      {
        key: "pitch",
        label: "Pay 1 life and exile a blue card from your hand",
        life: 1,
        pay_label: "a blue card from your hand",
        pay_options: { cards: ["brainstorm", "ponder"], min: 1, max: 1 },
        target_mode: "stack_spell",
      },
    ],
  });

  it("returns the payable cards for a cost that charges one", () => {
    expect(altCostPayOptions(alternativeCostByKey(forceOfWill, "pitch"))).toEqual([
      "brainstorm",
      "ponder",
    ]);
  });

  it("returns undefined for a cost that charges no cards", () => {
    // Overload charges mana only — the picker must be skipped, not
    // opened empty.
    expect(altCostPayOptions(alternativeCostByKey(rift, "overload"))).toBeUndefined();
    expect(altCostPayOptions(undefined)).toBeUndefined();
  });

  it("returns an empty array for a cost with nothing to pay it", () => {
    const stranded = card({
      instance_id: "fow2",
      alternative_costs: [{ key: "pitch", pay_options: { cards: [], min: 1, max: 1 } }],
    });
    expect(altCostPayOptions(alternativeCostByKey(stranded, "pitch"))).toEqual([]);
  });

  it("rides the cast payload as alt_cost_ids", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { altCost: "pitch", altCostIDs: ["brainstorm"] });
    expect(params).toEqual({ alternative_cost: "pitch", alt_cost_ids: ["brainstorm"] });
    // An empty list is omitted rather than sent: the server rejects a
    // non-empty one on a cost that charges nothing, and "I paid no
    // cards" is spelled by absence everywhere else in this payload.
    const empty: Record<string, unknown> = {};
    applyCastChoices(empty, { altCost: "overload", altCostIDs: [] });
    expect(empty).toEqual({ alternative_cost: "overload" });
  });
});

// S29 — escape is the first cost whose card-shaped half names more
// than one card, so the picker's count stopped being the constant 1
// it had been since S28. Everything about that lives in
// `pay_options.min`, and these are the three readings of it.
describe("escape's multi-card payment", () => {
  const cling = card({
    instance_id: "cling",
    name: "Cling to Dust",
    type_line: "Instant",
    mana_cost: "{B}",
    castable_here: true,
    alternative_costs: [
      {
        key: "escape",
        label: "Escape\u2014{3}{B}, Exile five other cards from your graveyard",
        mana_cost: "{3}{B}",
        pay_label: "five other cards from your graveyard",
        pay_options: { cards: ["a", "b", "c", "d", "e", "f"], min: 5, max: 5 },
      },
    ],
  });

  it("reads the count off pay_options", () => {
    expect(altCostPayCount(alternativeCostByKey(cling, "escape"))).toBe(5);
  });

  it("still reads one for the S28 single-card costs", () => {
    const pitcher = card({
      instance_id: "fow3",
      alternative_costs: [{ key: "pitch", pay_options: { cards: ["brainstorm"], min: 1, max: 1 } }],
    });
    expect(altCostPayCount(alternativeCostByKey(pitcher, "pitch"))).toBe(1);
  });

  // A cost with no card component never opens the picker, so the
  // count is never consulted — but it must not be zero, or a picker
  // opened by mistake would confirm with nothing chosen.
  it("falls back to one for an offer with no pay_options", () => {
    expect(altCostPayCount(alternativeCostByKey(rift, "overload"))).toBe(1);
    expect(altCostPayCount(undefined)).toBe(1);
  });

  it("sends every chosen card on the payload", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { altCost: "escape", altCostIDs: ["a", "b", "c", "d", "e"] });
    expect(params).toEqual({
      alternative_cost: "escape",
      alt_cost_ids: ["a", "b", "c", "d", "e"],
    });
  });
});
