import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import { begin, beginForModes, cancel, discardCostOf, targeting } from "./targeting";
import { canCastFromHand } from "./timing";
import type { CardView, GameView, ModeOptionView, PlayerView, ZoneView } from "./protocol";

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "p0", controller: "p0", ...extras };
}

describe("additional costs on cast — S21 sub-PR 5", () => {
  it("reads the discard count off the wire, and defaults to none", () => {
    expect(discardCostOf(card())).toBe(0);
    expect(discardCostOf(card({ additional_cost: {} }))).toBe(0);
    expect(
      discardCostOf(card({ additional_cost: { discard_cards: 1, label: "Discard a card" } })),
    ).toBe(1);
    expect(discardCostOf(card({ additional_cost: { discard_cards: 2 } }))).toBe(2);
  });

  it("carries the paid cards through a targeting prompt", () => {
    begin(card({ additional_cost: { discard_cards: 1 } }), "any", { discardIDs: ["fodder"] });
    expect(get(targeting)!.choices?.discardIDs).toEqual(["fodder"]);
    cancel();
  });

  it("carries them through a modal spell's targeting prompt too", () => {
    const option: ModeOptionView = {
      label: "target creature",
      target_mode: "creature",
      legal_targets: { cards: ["bear"], min: 1, max: 1 },
    };
    const modal = card({ modes: { prompt: "Choose one", min: 1, max: 1, options: [option] } });
    expect(beginForModes(modal, [0], { xValue: 3, discardIDs: ["fodder", "chaff"] })).toBe(true);
    const t = get(targeting)!;
    expect(t.choices?.discardIDs).toEqual(["fodder", "chaff"]);
    expect(t.choices?.xValue).toBe(3);
    expect(t.modes).toEqual([0]);
    cancel();
  });

  it("leaves discardIDs undefined when no cost was paid, so the payload omits it", () => {
    begin(card(), "any");
    expect(get(targeting)!.choices?.discardIDs).toBeUndefined();
    cancel();
  });
});

// A cost you can't pay makes the spell uncastable, and the spell
// itself never counts towards paying it.
describe("canCastFromHand — additional costs", () => {
  function zone(kind: string, cards: CardView[] = []): ZoneView {
    return { kind, owner: "p0", count: cards.length, cards };
  }

  // S31: the cast verdict is the server's enumerated move list, so
  // the fixture has to state what the server offered. `moves` is the
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

  const thrill = card({
    instance_id: "thrill",
    type_line: "Instant",
    additional_cost: { discard_cards: 1, label: "Discard a card" },
  });

  it("denies the cast when the hand holds only the spell", () => {
    // The server withheld it (CR 601.2h — the cost is unpayable) and
    // the client supplies the sentence.
    const verdict = canCastFromHand(thrill, snapshot([thrill]), "p0");
    expect(verdict.legal).toBe(false);
    expect(verdict.reason).toBe("No card to discard");
  });

  it("allows it once there is something else to pitch", () => {
    const fodder = card({ instance_id: "fodder", type_line: "Sorcery" });
    expect(canCastFromHand(thrill, snapshot([thrill, fodder], ["thrill"]), "p0").legal).toBe(true);
  });
});
