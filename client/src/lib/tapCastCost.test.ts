import { describe, expect, it } from "vitest";
import { applyCastChoices, castTapCost } from "./targeting";
import { castPreviewParams } from "./castPreview";
import { canCastFromHand } from "./timing";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";

// #3868: "tap an untapped Vampire you control" as a spell's additional
// cost (New Blood) — the claimed cost the cast chain asks about, its wire
// list, and the cast gate's sentence when nothing can pay.

const newBlood: CardView = {
  instance_id: "spell",
  name: "New Blood",
  owner: "p0",
  controller: "p0",
  type_line: "Sorcery",
  additional_cost: {
    label: "Tap an untapped Vampire you control",
    tap_others_label: "an untapped Vampire you control",
    tap_others_options: { cards: ["vamp", "count"], min: 1, max: 1 },
  },
};

describe("castTapCost", () => {
  it("reads the clause, its count and the permanents that could pay it", () => {
    expect(castTapCost(newBlood, {})).toEqual({
      label: "an untapped Vampire you control",
      count: 1,
      min: 1,
      options: ["vamp", "count"],
    });
  });

  // #3961: "tap any number of untapped creatures you control" ships min 0;
  // the picker may take none and up to every option.
  it("reads an any-number clause as min 0, up to every option", () => {
    const burn: CardView = {
      ...newBlood,
      name: "Burn at the Stake",
      additional_cost: {
        label: "Tap any number of untapped creatures you control",
        tap_others_label: "any number of untapped creatures you control",
        tap_others_options: { cards: ["a", "b", "c"], min: 0, max: 3 },
      },
    };
    expect(castTapCost(burn, {})).toEqual({
      label: "any number of untapped creatures you control",
      count: 3,
      min: 0,
      options: ["a", "b", "c"],
    });
  });

  it("is undefined for a card whose cost taps nothing", () => {
    expect(castTapCost({ ...newBlood, additional_cost: { label: "Discard a card" } }, {})).toBe(
      undefined,
    );
    expect(castTapCost({ ...newBlood, additional_cost: undefined }, {})).toBeUndefined();
  });

  it("reads the chosen either/or branch (Disruption Protocol)", () => {
    const protocol: CardView = {
      ...newBlood,
      name: "Disruption Protocol",
      additional_cost: {
        label: "Tap an untapped artifact you control or pay {1}",
        branches: [
          {
            key: "tap",
            label: "Tap an untapped artifact you control",
            tap_others_label: "an untapped artifact you control",
            tap_others_options: { cards: ["rock"], min: 1, max: 1 },
            payable: true,
          },
          { key: "mana", label: "Pay {1}", mana_cost: "{1}", payable: true },
        ],
      },
    };
    expect(castTapCost(protocol, { costBranch: 0 })?.options).toEqual(["rock"]);
    expect(castTapCost(protocol, { costBranch: 1 })).toBeUndefined();
  });
});

describe("the tapped permanents' wire list", () => {
  it("ride cast_spell as teamwork_ids, and the auto-tap preview keeps them out of the mana", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { teamworkIDs: ["vamp"] });
    expect(params.teamwork_ids).toEqual(["vamp"]);
    expect(castPreviewParams({ teamworkIDs: ["vamp"] }).teamworkIDs).toEqual(["vamp"]);
  });
});

describe("canCastFromHand — a tap cost nothing can pay", () => {
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
        seq: 1,
        number: 1,
        active_seat: 0,
        priority_holder: 0,
        phase: "precombat_main",
        step: "precombat_main",
      },
      mulligans_open: false,
      stack_items: [],
      split_second_active: false,
      legal_moves: [{ type: "pass_priority", player: "p0", kind: "pass", label: "Pass priority" }],
    };
  }

  it("says there is nothing to tap", () => {
    const none: CardView = {
      ...newBlood,
      additional_cost: {
        ...newBlood.additional_cost,
        tap_others_options: { cards: [], min: 1, max: 1 },
      },
    };
    const verdict = canCastFromHand(none, snapshot([none]), "p0");
    expect(verdict.legal).toBe(false);
    expect(verdict.reason).toBe("Nothing to tap");
  });

  it("names the count when the clause taps more than one", () => {
    const short: CardView = {
      ...newBlood,
      name: "Shared Discovery",
      additional_cost: { tap_others_options: { cards: ["a", "b", "c"], min: 4, max: 4 } },
    };
    expect(canCastFromHand(short, snapshot([short]), "p0").reason).toBe(
      "Needs 4 untapped permanents to tap",
    );
  });

  // #3961: "tap any number" is paid by tapping none, so an empty board is
  // never the reason a Burn at the Stake can't be cast.
  it("never blames an any-number cost", () => {
    const burn: CardView = {
      ...newBlood,
      name: "Burn at the Stake",
      additional_cost: { tap_others_options: { cards: [], min: 0, max: 0 } },
    };
    const reason = canCastFromHand(burn, snapshot([burn]), "p0").reason ?? "";
    expect(reason).not.toMatch(/tap/i);
  });
});
