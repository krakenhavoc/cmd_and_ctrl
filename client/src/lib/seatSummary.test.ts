import { describe, expect, it } from "vitest";
import type { CardView, PlayerView, ZoneView } from "./protocol";
import {
  attachmentHostIDs,
  buildSeatSummary,
  creatureStats,
  manaAvailable,
  manaLabel,
  structuralPermanents,
} from "./seatSummary";

// seatSummary.test.ts — the dense opponent read-out. The rules under
// test are the ones a player would act on: how much mana could answer
// them, what can block, and which permanents get a tile.
//
// The mana derivation is deliberately conservative. Several of these
// cases assert that something lands in `flexible` rather than in a
// colour — that is the contract, not a gap in it. A wrong colour is
// worse than an unknown one, because the player plays around it.

function zone(kind: string, owner: string | undefined, count: number): ZoneView {
  return { kind, owner, count, cards: [] };
}

function seat(id: string, over: Partial<PlayerView> = {}): PlayerView {
  return {
    id,
    name: id,
    seat: 0,
    life: 40,
    library: zone("library", id, 60),
    hand: zone("hand", id, 7),
    graveyard: zone("graveyard", id, 0),
    command: zone("command", id, 1),
    commander_damage: {},
    life_history: [],
    ...over,
  };
}

function card(over: Partial<CardView> & { instance_id: string }): CardView {
  return { name: over.instance_id, owner: "a", controller: "a", ...over };
}

function land(id: string, produced: string, over: Partial<CardView> = {}): CardView {
  return card({
    instance_id: id,
    type_line: "Land",
    mana_abilities: [{ index: 0, label: `Add ${produced}`, produced, tap_cost: true }],
    ...over,
  });
}

function creature(id: string, power: number, over: Partial<CardView> = {}): CardView {
  return card({ instance_id: id, type_line: "Creature — Human", power, toughness: power, ...over });
}

describe("manaAvailable", () => {
  it("counts an untapped basic as one mana of its colour", () => {
    const mana = manaAvailable([land("forest", "{G}")]);
    expect(mana.byColor).toEqual({ G: 1 });
    expect(mana.flexible).toBe(0);
    expect(mana.sources).toBe(1);
  });

  it("ignores a tapped source", () => {
    const mana = manaAvailable([land("forest", "{G}", { tapped: true })]);
    expect(mana.sources).toBe(0);
    expect(mana.byColor).toEqual({});
  });

  it("ignores a summoning-sick mana creature but not a sick mana rock", () => {
    // CR 302.6 binds the {T} cost, so a Bird that entered this turn
    // adds nothing. A Signet has no such restriction.
    const bird = card({
      instance_id: "bird",
      type_line: "Creature — Bird",
      summoning_sick: true,
      mana_abilities: [{ index: 0, produced: "{G}", tap_cost: true }],
    });
    const signet = card({
      instance_id: "signet",
      type_line: "Artifact",
      mana_abilities: [{ index: 0, produced: "{W}", tap_cost: false, mana_cost: "{1}" }],
    });
    const mana = manaAvailable([bird, signet]);
    expect(mana.byColor).toEqual({ W: 1 });
    expect(mana.sources).toBe(1);
  });

  it("adds two colourless for Temple of the False God's {C}{C}", () => {
    expect(manaAvailable([land("temple", "{C}{C}")]).byColor).toEqual({ C: 2 });
  });

  it("greys out a mana ability whose condition is unmet", () => {
    // Temple of the False God with fewer than five lands.
    const temple = card({
      instance_id: "temple",
      type_line: "Land",
      mana_abilities: [{ index: 0, produced: "{C}{C}", tap_cost: true, condition_unmet: true }],
    });
    expect(manaAvailable([temple]).sources).toBe(0);
  });

  it("drops a source an effect has shut off", () => {
    const shut = land("shut", "{G}", { restrictions: ["cant_activate_mana"] });
    expect(manaAvailable([shut]).sources).toBe(0);
  });

  it("drops a Command Tower with no commander identity (CR 903.4f)", () => {
    const tower = card({
      instance_id: "tower",
      type_line: "Land",
      mana_abilities: [{ index: 0, produced: "", tap_cost: true, adds_no_mana: true }],
    });
    expect(manaAvailable([tower]).sources).toBe(0);
  });

  it("counts a dual land once, not once per ability", () => {
    // A Volcanic Island can only be tapped once. Summing its printed
    // options would report two mana from one land.
    const dual = card({
      instance_id: "dual",
      type_line: "Land",
      mana_abilities: [
        { index: 0, produced: "{U}", tap_cost: true },
        { index: 1, produced: "{R}", tap_cost: true },
      ],
    });
    const mana = manaAvailable([dual]);
    expect(mana.sources).toBe(1);
    expect(Object.values(mana.byColor).reduce((a, b) => a + b, 0)).toBe(1);
  });

  it("prefers a nameable colour when the abilities are mixed", () => {
    // The flexible entry is listed first; the named one should still
    // be what gets reported, rather than the land vanishing into
    // `flexible` on listing order alone.
    const mixed = card({
      instance_id: "mixed",
      type_line: "Land",
      mana_abilities: [
        { index: 0, produced: "any color", tap_cost: true },
        { index: 1, produced: "{B}", tap_cost: true },
      ],
    });
    expect(manaAvailable([mixed]).byColor).toEqual({ B: 1 });
  });

  it("puts an any-colour source in flexible, never in a colour", () => {
    const birds = card({
      instance_id: "birds",
      type_line: "Creature — Bird",
      mana_abilities: [{ index: 0, produced: "any color", tap_cost: true }],
    });
    const mana = manaAvailable([birds]);
    expect(mana.byColor).toEqual({});
    expect(mana.flexible).toBe(1);
    expect(mana.sources).toBe(1);
  });

  it("puts a hybrid symbol in flexible", () => {
    const mana = manaAvailable([land("hybrid", "{W/U}")]);
    expect(mana.byColor).toEqual({});
    expect(mana.flexible).toBe(1);
  });

  it("treats a generic amount as that many flexible mana", () => {
    // Cabal Coffers pays out an amount, not a colour.
    const mana = manaAvailable([land("coffers", "{3}")]);
    expect(mana.byColor).toEqual({});
    expect(mana.flexible).toBe(3);
  });

  it("treats an absent produced string as one flexible mana", () => {
    const odd = card({
      instance_id: "odd",
      type_line: "Artifact",
      mana_abilities: [{ index: 0, label: "Add mana", tap_cost: true }],
    });
    expect(manaAvailable([odd]).flexible).toBe(1);
  });

  it("sums across a real-ish board", () => {
    const mana = manaAvailable([
      land("f1", "{G}"),
      land("f2", "{G}"),
      land("i1", "{U}"),
      land("t1", "{G}", { tapped: true }),
      land("sol", "{C}{C}"),
    ]);
    expect(mana.byColor).toEqual({ G: 2, U: 1, C: 2 });
    expect(mana.sources).toBe(4);
  });

  it("reports nothing for a board with no mana sources", () => {
    const mana = manaAvailable([creature("bear", 2)]);
    expect(mana).toEqual({ byColor: {}, flexible: 0, sources: 0 });
  });
});

describe("manaLabel", () => {
  it("says nothing rather than zero for an empty board", () => {
    expect(manaLabel({ byColor: {}, flexible: 0, sources: 0 })).toBe("no mana");
  });

  it("counts flexible mana into the total and names the source count", () => {
    expect(manaLabel({ byColor: { G: 2 }, flexible: 1, sources: 3 })).toBe("3 mana from 3 sources");
  });

  it("uses the singular for one source", () => {
    expect(manaLabel({ byColor: { U: 1 }, flexible: 0, sources: 1 })).toBe("1 mana from 1 source");
  });
});

describe("creatureStats", () => {
  it("splits tapped from untapped and sums power", () => {
    const stats = creatureStats([
      creature("a", 2),
      creature("b", 3),
      creature("c", 5, { tapped: true }),
      land("l", "{G}"),
    ]);
    expect(stats).toEqual({ total: 3, untapped: 2, tapped: 1, power: 10, untappedPower: 5 });
  });

  it("counts a creature with no numeric power as a body worth zero", () => {
    const stats = creatureStats([creature("x", 0, { power: undefined })]);
    expect(stats.total).toBe(1);
    expect(stats.power).toBe(0);
  });

  it("counts an artifact creature as a creature", () => {
    const golem = card({ instance_id: "g", type_line: "Artifact Creature — Golem", power: 4 });
    expect(creatureStats([golem]).total).toBe(1);
  });

  it("is all zeroes for an empty board", () => {
    expect(creatureStats([])).toEqual({
      total: 0,
      untapped: 0,
      tapped: 0,
      power: 0,
      untappedPower: 0,
    });
  });
});

describe("attachmentHostIDs", () => {
  it("names the permanent an Aura sits on", () => {
    const bear = creature("bear", 2);
    const aura = card({
      instance_id: "aura",
      type_line: "Enchantment — Aura",
      attached_to: { kind: "card", id: "bear" },
    });
    expect([...attachmentHostIDs([bear, aura])]).toEqual(["bear"]);
  });

  it("ignores a dangling attachment whose host has left", () => {
    const aura = card({
      instance_id: "aura",
      type_line: "Enchantment — Aura",
      attached_to: { kind: "card", id: "gone" },
    });
    expect(attachmentHostIDs([aura]).size).toBe(0);
  });

  it("ignores a Curse, which enchants a player", () => {
    const curse = card({
      instance_id: "curse",
      type_line: "Enchantment — Aura Curse",
      attached_to: { kind: "player", id: "b" },
    });
    expect(attachmentHostIDs([curse]).size).toBe(0);
  });
});

describe("structuralPermanents", () => {
  it("picks planeswalkers with their loyalty", () => {
    const pw = card({
      instance_id: "jace",
      type_line: "Legendary Planeswalker — Jace",
      counters: { loyalty: 4 },
    });
    expect(structuralPermanents([pw], new Set())).toEqual([
      { card: pw, kind: "planeswalker", counter: 4 },
    ]);
  });

  it("picks the commander even though it is a creature", () => {
    const cmdr = creature("cmdr", 3, { is_commander: true });
    expect(structuralPermanents([cmdr], new Set())[0]?.kind).toBe("commander");
  });

  it("picks an enchanted creature", () => {
    const bear = creature("bear", 2);
    expect(structuralPermanents([bear], new Set(["bear"]))[0]?.kind).toBe("attached");
  });

  it("leaves an ordinary creature and an ordinary land alone", () => {
    expect(structuralPermanents([creature("bear", 2), land("f", "{G}")], new Set())).toEqual([]);
  });

  it("does not rank or truncate — every qualifying permanent is returned", () => {
    const walkers = [0, 1, 2, 3, 4].map((i) =>
      card({ instance_id: `pw${i}`, type_line: "Planeswalker — Test", counters: { loyalty: i } }),
    );
    expect(structuralPermanents(walkers, new Set())).toHaveLength(5);
  });
});

describe("buildSeatSummary", () => {
  it("assembles the read-out for a live seat", () => {
    const board = [land("f", "{G}"), creature("bear", 2)];
    const s = buildSeatSummary(seat("a", { life: 33 }), board, board);
    expect(s.life).toBe(33);
    expect(s.handCount).toBe(7);
    expect(s.mana.byColor).toEqual({ G: 1 });
    expect(s.creatures.total).toBe(1);
    expect(s.creatureCards.map((c) => c.instance_id)).toEqual(["bear"]);
    expect(s.eliminated).toBe(false);
  });

  it("drops zero entries from commander damage", () => {
    const s = buildSeatSummary(seat("a", { commander_damage: { x: 0, y: 7 } }), [], []);
    expect(s.commanderDamage).toEqual({ y: 7 });
  });

  it("reports an eliminated seat as empty rather than as a board of zeroes", () => {
    const board = [land("f", "{G}"), creature("bear", 2)];
    const s = buildSeatSummary(seat("a", { eliminated: true }), board, board);
    expect(s.eliminated).toBe(true);
    expect(s.mana.sources).toBe(0);
    expect(s.creatures.total).toBe(0);
    expect(s.handCount).toBe(0);
    // Identity and life survive: the seat is still at the table.
    expect(s.life).toBe(40);
  });

  it("sees an attachment controlled by another seat", () => {
    // CR 303.4f — an Aura you control can enchant a creature an
    // opponent controls, so the host set has to come off the whole
    // battlefield rather than one seat's slice.
    const bear = creature("bear", 2, { controller: "a" });
    const aura = card({
      instance_id: "aura",
      controller: "b",
      type_line: "Enchantment — Aura",
      attached_to: { kind: "card", id: "bear" },
    });
    const s = buildSeatSummary(seat("a"), [bear], [bear, aura]);
    expect(s.structural.map((p) => p.kind)).toEqual(["attached"]);
  });
});
