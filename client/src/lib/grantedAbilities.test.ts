// ADR 0093 on the client: granted ability rows carry a `ref` the
// activate verbs send back, a left-click on a permanent with a granted
// row never picks silently, and the menu names the grantor.

import { describe, it, expect } from "vitest";

import {
  activatedAbilityRef,
  grantedFromLabel,
  hasGrantedActivatedAbility,
  hasGrantedAbility,
  isStaleAbilityRefError,
  manaAbilityRef,
} from "./abilityRef";
import { battlefieldClickIntent, buildMenuSections } from "./contextMenu.logic";
import { manaAbilityOptions, manaClickPlan } from "./manaSource";
import type { CardView, GameView, ManaAbilityView } from "./protocol";

const card = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "c1",
    name: "Bear",
    owner: "me",
    controller: "me",
    type_line: "Creature — Bear",
    ...extra,
  }) as CardView;

const rite = { id: "rite", name: "Cryptolith Rite" };

const grantedAnyColor = (index: number): ManaAbilityView => ({
  index,
  ref: `grant:cryptolith-rite/any-color:0:0`,
  granted_by: rite,
  tap_cost: true,
  produced: "{W|U|B|R|G}",
  label: "Add one mana of any color",
});

// A creature under Cryptolith Rite: its only mana row is granted.
const bearUnderRite = (): CardView => card({ mana_abilities: [grantedAnyColor(0)] });

// Jaheira's grant on a token: ONE fixed-output granted row.
const tokenUnderJaheira = (): CardView =>
  card({
    name: "Goblin",
    mana_abilities: [
      {
        index: 0,
        ref: "grant:jaheira-friend-of-the-forest/tap-for-green:0:0",
        granted_by: { id: "jaheira", name: "Jaheira, Friend of the Forest" },
        tap_cost: true,
        produced: "{G}",
        label: "Add {G}",
      },
    ],
  });

// A Forest under Chromatic Lantern: its own {G} and the granted any colour.
const forestUnderLantern = (): CardView =>
  card({
    name: "Forest",
    type_line: "Basic Land — Forest",
    mana_abilities: [
      { index: 0, ref: "land:G", tap_cost: true, produced: "{G}", label: "Add {G}" },
      { ...grantedAnyColor(1), granted_by: { id: "lantern", name: "Chromatic Lantern" } },
    ],
  });

// A land enchanted by Squirrel Nest: a mana ability of its own and a
// granted activated ability.
const nestLand = (): CardView =>
  card({
    name: "Forest",
    type_line: "Basic Land — Forest",
    mana_abilities: [{ index: 0, ref: "land:G", tap_cost: true, produced: "{G}" }],
    activated_abilities: [
      {
        index: 0,
        ref: "grant:squirrel-nest/make-a-squirrel:0:0",
        granted_by: { id: "nest", name: "Squirrel Nest" },
        label: "{T}: Create a 1/1 green Squirrel creature token",
        tap_cost: true,
      },
    ],
  });

describe("ability refs", () => {
  it("reads the ref of the row at the index, and nothing for a missing row", () => {
    expect(manaAbilityRef(forestUnderLantern(), 1)).toEqual({
      ref: "grant:cryptolith-rite/any-color:0:0",
    });
    expect(manaAbilityRef(forestUnderLantern(), 0)).toEqual({ ref: "land:G" });
    expect(manaAbilityRef(forestUnderLantern(), 7)).toEqual({});
    expect(activatedAbilityRef(nestLand(), 0)).toEqual({
      ref: "grant:squirrel-nest/make-a-squirrel:0:0",
    });
  });

  it("sends nothing when an older server published no ref", () => {
    expect(manaAbilityRef(card({ mana_abilities: [{ index: 0, produced: "{G}" }] }), 0)).toEqual(
      {},
    );
  });

  it("recognises the stale-ref refusal", () => {
    expect(
      isStaleAbilityRefError(
        "game: that ability is no longer at that position — the board changed",
      ),
    ).toBe(true);
    expect(isStaleAbilityRefError("game: card already tapped")).toBe(false);
    expect(isStaleAbilityRefError(undefined)).toBe(false);
  });

  it("knows which permanents carry a granted row", () => {
    expect(hasGrantedAbility(bearUnderRite())).toBe(true);
    expect(hasGrantedAbility(card())).toBe(false);
    expect(hasGrantedActivatedAbility(nestLand())).toBe(true);
    expect(hasGrantedActivatedAbility(bearUnderRite())).toBe(false);
    expect(grantedFromLabel({ granted_by: rite })).toBe("from Cryptolith Rite");
    expect(grantedFromLabel({})).toBe("");
  });
});

describe("left-click on a permanent with a granted ability (owner decision, 2026-09-24)", () => {
  const mana = { manaClick: true };

  it("clicks a creature under Cryptolith Rite for mana, through the picker", () => {
    expect(battlefieldClickIntent(bearUnderRite(), "me", false, mana)).toBe("mana");
    expect(manaClickPlan(bearUnderRite())?.kind).toBe("pick");
  });

  it("never activates a lone granted ability silently", () => {
    expect(manaClickPlan(tokenUnderJaheira())?.kind).toBe("pick");
  });

  it("opens the picker for a Forest under Chromatic Lantern", () => {
    expect(manaClickPlan(forestUnderLantern())?.kind).toBe("pick");
  });

  it("opens the ability menu for a land enchanted by Squirrel Nest", () => {
    expect(battlefieldClickIntent(nestLand(), "me", false, mana)).toBe("abilities");
  });

  it("labels a granted picker option with its grantor", () => {
    const opts = manaAbilityOptions(tokenUnderJaheira());
    expect(opts).toHaveLength(1);
    expect(opts[0].granted).toBe("from Jaheira, Friend of the Forest");
    expect(opts[0].rider).toContain("from Jaheira, Friend of the Forest");
  });
});

describe("the card menu names the grantor", () => {
  const view = (c: CardView): GameView =>
    ({
      seats: [{ id: "me", name: "Me" }],
      battlefield: { kind: "battlefield", count: 1, cards: [c] },
      stack: { kind: "stack", count: 0, cards: [] },
      exile: { kind: "exile", count: 0, cards: [] },
      turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    }) as unknown as GameView;

  it("suffixes a granted row with (from <grantor>) and leaves an own row alone", () => {
    const labels = buildMenuSections(view(nestLand()), nestLand(), "me", false)
      .flatMap((s) => s.items)
      .map((i) => i.label);
    expect(labels).toContain(
      "{T}: Create a 1/1 green Squirrel creature token (from Squirrel Nest)",
    );
    expect(labels.some((l) => l.includes("(from") && l.includes("Add {G}"))).toBe(false);
  });
});
