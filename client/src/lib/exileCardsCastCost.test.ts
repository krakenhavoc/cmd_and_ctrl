import { describe, expect, it } from "vitest";
import { applyCastChoices, castExileCardsCost } from "./targeting";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import { exileCostNote, exileCostOptionCards, exileCostWhere } from "./exileCost";
import { canCastFromHand } from "./timing";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";

// #3536: "exile six cards from your graveyard" as a spell's additional cost
// (Abhorrent Oculus) — the claimed cost the cast chain asks about (the
// graveyard picker), the one wire list, and the cast gate.

const six = ["g1", "g2", "g3", "g4", "g5", "g6"];

function oculus(options: string[]): CardView {
  return {
    instance_id: "spell",
    name: "Abhorrent Oculus",
    owner: "me",
    controller: "me",
    additional_cost: {
      label: "Exile six cards from your graveyard",
      exile_cost_n: 6,
      exile_cost_label: "six cards",
      exile_cost_options: options,
      exile_cost_zone: "graveyard",
    },
  };
}

describe("castExileCardsCost", () => {
  it("reads the clause, its count and the graveyard cards that could pay it", () => {
    const got = castExileCardsCost(oculus([...six, "g7"]), {});
    expect(got?.label).toBe("six cards");
    expect(got?.count).toBe(6);
    expect(got?.options).toEqual([...six, "g7"]);
    expect(got?.cost.exile_cost_zone).toBe("graveyard");
  });

  it("is undefined for a card whose cost exiles no cards", () => {
    const plain: CardView = { ...oculus(six), additional_cost: { label: "Sacrifice a creature" } };
    expect(castExileCardsCost(plain, {})).toBeUndefined();
    expect(castExileCardsCost({ ...oculus(six), additional_cost: undefined }, {})).toBeUndefined();
  });

  it("opens the picker over the graveyard, not the hand", () => {
    const seat = {
      id: "me",
      hand: { cards: [{ instance_id: "h1", name: "Hand Card" }] },
      graveyard: { cards: [...six, "g7"].map((id) => ({ instance_id: id, name: id })) },
    } as unknown as PlayerView;
    const cost = castExileCardsCost(oculus([...six, "g7"]), {})!.cost;
    expect(exileCostOptionCards(seat, cost).map((c) => c.instance_id)).toEqual([...six, "g7"]);
    expect(exileCostWhere(cost)).toBe("in your graveyard");
    expect(exileCostNote(cost)).toContain("graveyard");
  });
});

describe("the exile_ids wire list", () => {
  it("rides cast_spell, omitted when empty", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { exileIDs: six });
    expect(params.exile_ids).toEqual(six);
    const empty: Record<string, unknown> = {};
    applyCastChoices(empty, { exileIDs: [] });
    expect(empty).not.toHaveProperty("exile_ids");
  });

  it("stays off the auto-tap preview: a graveyard card is no mana source", () => {
    expect(castPreviewParams({ exileIDs: six })).not.toHaveProperty("exileIDs");
    expect(castPreviewParamsFromPayload({ exile_ids: six })).not.toHaveProperty("exileIDs");
  });
});

describe("the cast gate", () => {
  function zone(kind: string, cards: CardView[] = []): ZoneView {
    return { kind, owner: "p0", count: cards.length, cards };
  }

  // The server's cast verdict is the enumerated move list: `offered` says
  // whether it listed a cast for the card.
  function snapshot(hand: CardView[], offered: boolean): GameView {
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
      legal_moves: [
        { type: "pass_priority", player: "p0", kind: "pass", label: "Pass priority" },
        ...(offered
          ? [
              {
                type: "cast_spell",
                player: "p0",
                kind: "cast" as const,
                label: "Cast Abhorrent Oculus",
                source: "spell",
              },
            ]
          : []),
      ],
    };
  }

  it("greys a spell whose graveyard holds fewer cards than the cost demands", () => {
    const card = oculus(six.slice(0, 5));
    const verdict = canCastFromHand(card, snapshot([card], false), "p0");
    expect(verdict.legal).toBe(false);
    expect(verdict.reason).toBe("Needs 6 cards in your graveyard");
  });

  it("does not blame the graveyard with exactly six, or with more", () => {
    for (const options of [six, [...six, "g7"]]) {
      const card = oculus(options);
      expect(canCastFromHand(card, snapshot([card], true), "p0").legal).toBe(true);
      // Withheld for some other reason (mana, timing): not this clause.
      expect(canCastFromHand(card, snapshot([card], false), "p0").reason).not.toContain(
        "graveyard",
      );
    }
  });
});
