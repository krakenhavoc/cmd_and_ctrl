import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import {
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

  function snapshot(hand: CardView[]): GameView {
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
    };
  }

  it("stays castable with an empty printed legal set when a cost clears the clause", () => {
    expect(canCastFromHand(rift, snapshot([rift]), "p0").legal).toBe(true);
  });

  it("still denies a card whose every cost option has nothing to point at", () => {
    const verdict = canCastFromHand(washAway, snapshot([washAway]), "p0");
    // The cleave clause has a legal spell, so this one is castable —
    // flip the cleave set empty and the denial comes back.
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

  it("leaves a card with no alternative costs on the old gate", () => {
    const bolt = card({
      instance_id: "bolt",
      type_line: "Instant",
      legal_targets: { cards: [], min: 1, max: 1 },
    });
    expect(canCastFromHand(bolt, snapshot([bolt]), "p0").reason).toBe("No legal target");
  });
});
