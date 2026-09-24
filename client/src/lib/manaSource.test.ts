// #1438: a left-click on a mana source taps it FOR mana. The routing
// decision and the picker's options are pure, so they are pinned here;
// manaClick.render.test.ts checks that the clicks reach them.

import { describe, it, expect } from "vitest";

import { battlefieldClickIntent, buildMenuSections } from "./contextMenu.logic";
import { colorButtons } from "./manaPick";
import {
  colorCombos,
  colorPickOptions,
  labelRider,
  manaAbilityOptions,
  manaClickPlan,
  manaColorParams,
  placePopover,
} from "./manaSource";
import { MANA_SYMBOL_META, manaSymbolMeta } from "./manaSymbol";
import type { CardView, GameView, ManaAbilityView } from "./protocol";

const card = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "c1",
    name: "Swamp",
    owner: "me",
    controller: "me",
    type_line: "Basic Land — Swamp",
    ...extra,
  }) as CardView;

const ability = (index: number, extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index,
  tap_cost: true,
  ...extra,
});

const swamp = (): CardView =>
  card({ mana_abilities: [ability(0, { produced: "{B}", label: "Add {B}" })] });

// Battlefield Forge: the painless {C} at index 0, the pain dual at 1
// (server/internal/cards/effects/mana_rider_helpers.go).
const forge = (): CardView =>
  card({
    instance_id: "forge",
    name: "Battlefield Forge",
    type_line: "Land",
    mana_abilities: [
      ability(0, { produced: "{C}", label: "Add {C}" }),
      ability(1, {
        produced: "{R|W}",
        label: "Add {R} or {W}. This land deals 1 damage to you.",
      }),
    ],
  });

describe("manaClickPlan", () => {
  it("activates a single fixed-output ability at once (Swamp)", () => {
    expect(manaClickPlan(swamp())).toEqual({ kind: "activate", index: 0 });
  });

  it("activates Sol Ring's {C}{C} at once", () => {
    const solRing = card({
      name: "Sol Ring",
      type_line: "Artifact",
      mana_abilities: [ability(0, { produced: "{C}{C}" })],
    });
    expect(manaClickPlan(solRing)).toEqual({ kind: "activate", index: 0 });
  });

  it("activates a single any-colour ability at once when the server publishes no colours", () => {
    // A server from before #1443 publishes no color_options: the
    // colours are then its mana_pick prompt after the tap, narrowed
    // and ordered there (#843), and the client guesses nothing.
    const birds = card({
      name: "Birds of Paradise",
      type_line: "Creature — Bird",
      mana_abilities: [ability(0, { produced: "{W|U|B|R|G}" })],
    });
    expect(manaClickPlan(birds)).toEqual({ kind: "activate", index: 0 });
  });

  it("does not pre-empt summoning sickness — the server refuses it", () => {
    const sick = card({
      type_line: "Creature — Elf Druid",
      summoning_sick: true,
      mana_abilities: [ability(0, { produced: "{G}" })],
    });
    expect(manaClickPlan(sick)).toEqual({ kind: "activate", index: 0 });
  });

  it("opens the picker for a source with several abilities, in the server's order", () => {
    const plan = manaClickPlan(forge());
    expect(plan?.kind).toBe("pick");
    if (plan?.kind !== "pick") return;
    expect(plan.options.map((o) => o.abilityIndex)).toEqual([0, 1]);
    expect(plan.options[0]).toMatchObject({ symbols: ["C"], caption: "Colorless" });
    expect(plan.options[0].rider).toBeUndefined();
    expect(plan.options[1]).toMatchObject({
      symbols: ["R", "W"],
      choice: true,
      caption: "Red or White",
      rider: "deals 1 damage to you",
    });
  });

  it("opens the picker to say why when the only ability is greyed by the server", () => {
    const tower = card({
      name: "Command Tower",
      type_line: "Land",
      mana_abilities: [ability(0, { produced: "{W|U|B|R|G}", adds_no_mana: true })],
    });
    const plan = manaClickPlan(tower);
    expect(plan?.kind).toBe("pick");
    if (plan?.kind !== "pick") return;
    expect(plan.options[0].disabled).toMatch(/commander/);
  });

  it("has no plan for a permanent without mana abilities", () => {
    expect(manaClickPlan(card({ type_line: "Creature — Bear" }))).toBeNull();
  });
});

describe("mana ability options", () => {
  it("lists every cost that is not {T} as a rider", () => {
    const [o] = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, { produced: "{W|U|B|R|G}", life_cost: 1, label: "Add one mana of any color" }),
        ],
      }),
    );
    expect(o.rider).toBe("pay 1 life");
    const [petal] = manaAbilityOptions(
      card({ mana_abilities: [ability(0, { produced: "{W|U|B|R|G}", sacrifice_cost: true })] }),
    );
    expect(petal.rider).toBe("sacrifice it");
    const [signet] = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, { produced: "{W}{U}", mana_cost: "{1}", charged_mana_cost: "{1}" }),
        ],
      }),
    );
    expect(signet).toMatchObject({
      symbols: ["W", "U"],
      caption: "White and Blue",
      rider: "pay {1}",
    });
  });

  it("calls a five-colour alternative 'Any color' (Tarnished Citadel)", () => {
    const [, o] = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, { produced: "{C}" }),
          ability(1, {
            produced: "{W|U|B|R|G}",
            label: "Add one mana of any color. This land deals 3 damage to you",
          }),
        ],
      }),
    );
    expect(o).toMatchObject({ caption: "Any color", rider: "deals 3 damage to you" });
  });

  it("counts a repeated symbol in the caption", () => {
    const [o] = manaAbilityOptions(card({ mana_abilities: [ability(0, { produced: "{C}{C}" })] }));
    expect(o.caption).toBe("2 Colorless");
  });

  it("falls back to the server's label for a computed output", () => {
    const [o] = manaAbilityOptions(
      card({ mana_abilities: [ability(0, { label: "Add {B} for each Swamp you control" })] }),
    );
    expect(o.symbols).toEqual([]);
    expect(o.caption).toBe("Add {B} for each Swamp you control");
  });

  it("reads the painland rider out of the label", () => {
    expect(labelRider("Add {C}{C}. This land deals 2 damage to you.")).toBe(
      "deals 2 damage to you",
    );
    expect(labelRider("Add {B}. This creature deals 1 damage to you.")).toBe(
      "deals 1 damage to you",
    );
    expect(labelRider("Add {C}")).toBe("");
    expect(labelRider(undefined)).toBe("");
  });

  it("greys an option the server flagged, and only those", () => {
    const opts = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, { produced: "{C}" }),
          ability(1, { produced: "{G}", condition_unmet: true }),
          ability(2, { produced: "{G}{G}{G}", exhausted: true }),
        ],
      }),
    );
    expect(opts.map((o) => !!o.disabled)).toEqual([false, true, true]);
  });
});

describe("colorPickOptions (the mana_pick prompt)", () => {
  it("keeps the server's order exactly — commander identity first (#843)", () => {
    const opts = colorPickOptions(colorButtons(["G", "B", "W", "U", "R"]), "add");
    expect(opts.map((o) => o.color)).toEqual(["G", "B", "W", "U", "R"]);
    expect(opts[0]).toMatchObject({ symbols: ["G"], caption: "Green", title: "Add Green mana" });
  });

  it("draws an N-mana pick as N symbols (#742)", () => {
    const [o] = colorPickOptions(colorButtons(["R"], { R: 3 }), "add");
    expect(o.symbols).toEqual(["R", "R", "R"]);
    expect(o.caption).toBe("3 Red");
    expect(o.title).toBe("Add 3 Red mana");
  });

  it("words a choose_color option as a choice, not mana", () => {
    const [o] = colorPickOptions(colorButtons(["U"]), "choose");
    expect(o.title).toBe("Choose Blue");
  });
});

describe("mana symbols", () => {
  it("has a glyph and a disc colour for all five colours and colourless", () => {
    for (const s of ["W", "U", "B", "R", "G", "C"]) {
      expect(MANA_SYMBOL_META[s].glyph.length, s).toBeGreaterThan(10);
    }
    expect(manaSymbolMeta("C").name).toBe("Colorless");
  });

  it("falls back to a grey text disc for a generic symbol", () => {
    expect(manaSymbolMeta("2")).toMatchObject({ name: "2", glyph: "" });
  });
});

describe("placePopover", () => {
  const land = { left: 100, top: 600, right: 170, bottom: 700 };

  it("prefers above the card", () => {
    expect(placePopover(land, 200, 120, 1280, 800)).toEqual({
      left: 35,
      top: 472,
      side: "above",
    });
  });

  it("goes below when there is no room above", () => {
    const top = { left: 100, top: 20, right: 170, bottom: 120 };
    expect(placePopover(top, 200, 120, 1280, 800).side).toBe("below");
  });

  it("stays inside a 390 px phone with the 16 px gutter", () => {
    const right = { left: 340, top: 600, right: 385, bottom: 680 };
    const p = placePopover(right, 300, 150, 390, 844);
    expect(p.left).toBe(390 - 16 - 300);
    const left = placePopover({ left: 0, top: 600, right: 40, bottom: 680 }, 300, 150, 390, 844);
    expect(left.left).toBe(16);
  });

  it("pins to the gutter when the box is wider than the phone", () => {
    expect(placePopover(land, 500, 100, 390, 844).left).toBe(16);
  });
});

describe("battlefieldClickIntent — #1438 click for mana", () => {
  const mana = { manaClick: true };

  it("clicks an untapped mana source for mana", () => {
    expect(battlefieldClickIntent(swamp(), "me", false, mana)).toBe("mana");
    expect(battlefieldClickIntent(forge(), "me", false, mana)).toBe("mana");
  });

  it("untaps a tapped source with one click", () => {
    const tapped = { ...swamp(), tapped: true };
    expect(battlefieldClickIntent(tapped, "me", false, mana)).toBe("tap");
  });

  it("raw-taps on Alt-click", () => {
    expect(battlefieldClickIntent(swamp(), "me", false, { ...mana, rawTap: true })).toBe("tap");
  });

  it("keeps click-to-tap for a permanent with no mana ability", () => {
    const bear = card({ type_line: "Creature — Bear" });
    expect(battlefieldClickIntent(bear, "me", false, mana)).toBe("tap");
  });

  it("raw-taps where the panel cannot activate (an admin on another seat's land)", () => {
    expect(battlefieldClickIntent(swamp(), "admin", true)).toBe("tap");
  });

  it("clicks a utility land that also makes mana for its mana (Rogue's Passage)", () => {
    const passage = card({
      name: "Rogue's Passage",
      type_line: "Land",
      mana_abilities: [ability(0, { produced: "{C}" })],
      activated_abilities: [{ index: 0, label: "{4}, {T}: can't be blocked" }],
    } as Partial<CardView>);
    expect(battlefieldClickIntent(passage, "me", false, mana)).toBe("mana");
  });

  it("still opens a planeswalker's menu (#329)", () => {
    const walker = card({
      type_line: "Legendary Planeswalker — Nissa",
      mana_abilities: [ability(0, { produced: "{G}" })],
    });
    expect(battlefieldClickIntent(walker, "me", false, mana)).toBe("abilities");
  });

  it("does nothing on an opponent's source", () => {
    expect(battlefieldClickIntent(swamp(), "someone-else", false, mana)).toBe("none");
  });
});

describe("the override menu's raw tap", () => {
  const view = (c: CardView): GameView =>
    ({
      seats: [{ id: "me", name: "Me" }],
      battlefield: { kind: "battlefield", count: 1, cards: [c] },
      stack: { kind: "stack", count: 0, cards: [] },
      exile: { kind: "exile", count: 0, cards: [] },
      turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    }) as unknown as GameView;

  const tapLabel = (c: CardView): string | undefined =>
    buildMenuSections(view(c), c, "me", false)
      .flatMap((s) => s.items)
      .find((i) => i.id === "tap")?.label;

  it("says 'Tap (no mana)' on a mana source", () => {
    expect(tapLabel(swamp())).toBe("Tap (no mana)");
  });

  it("stays 'Tap' on a permanent with no mana ability", () => {
    expect(tapLabel(card({ type_line: "Creature — Bear" }))).toBe("Tap");
  });
});

// #1443: the server publishes each ability's color_options (the list
// the mana_pick would carry), so the picker offers FINAL results and
// the colour rides the activation.
describe("colour chosen before the tap (#1443)", () => {
  const forgeWithColors = (): CardView =>
    card({
      instance_id: "forge",
      name: "Battlefield Forge",
      type_line: "Land",
      mana_abilities: [
        ability(0, { produced: "{C}", label: "Add {C}" }),
        ability(1, {
          produced: "{R|W}",
          label: "Add {R} or {W}. This land deals 1 damage to you.",
          color_options: [["R", "W"]],
        }),
      ],
    });

  it("expands a painland into {C}, {R} and {W}, the coloured two with the damage", () => {
    const plan = manaClickPlan(forgeWithColors());
    expect(plan?.kind).toBe("pick");
    if (plan?.kind !== "pick") return;
    expect(plan.options.map((o) => [o.abilityIndex, o.symbols, o.colors, o.rider])).toEqual([
      [0, ["C"], undefined, undefined],
      [1, ["R"], ["R"], "deals 1 damage to you"],
      [1, ["W"], ["W"], "deals 1 damage to you"],
    ]);
    expect(plan.options[1]).toMatchObject({
      caption: "Red",
      title: "Add {R} — deals 1 damage to you",
    });
    expect(plan.options[1].choice).toBeFalsy();
  });

  it("keeps the server's colour order for Birds of Paradise (#843)", () => {
    const opts = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, { produced: "{W|U|B|R|G}", color_options: [["G", "W", "U", "B", "R"]] }),
        ],
      }),
    );
    expect(opts.map((o) => o.colors)).toEqual([["G"], ["W"], ["U"], ["B"], ["R"]]);
  });

  it("taps a one-colour answer at once and names the colour (mono-green Command Tower)", () => {
    const tower = card({
      name: "Command Tower",
      type_line: "Land",
      mana_abilities: [ability(0, { produced: "{W|U|B|R|G}", color_options: [["G"]] })],
    });
    expect(manaClickPlan(tower)).toEqual({ kind: "activate", index: 0, colors: ["G"] });
  });

  it("offers each distinct result of a two-slot filter land once (Mystic Gate)", () => {
    const [, ...opts] = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, { produced: "{C}" }),
          ability(1, {
            produced: "{W|U}{W|U}",
            mana_cost: "{W/U}",
            color_options: [
              ["W", "U"],
              ["W", "U"],
            ],
          }),
        ],
      }),
    );
    expect(opts.map((o) => o.colors)).toEqual([
      ["W", "W"],
      ["W", "U"],
      ["U", "U"],
    ]);
    expect(opts[1]).toMatchObject({ symbols: ["W", "U"], caption: "White and Blue" });
    expect(opts[0].rider).toBe("pay {W/U}");
  });

  it("draws an 'N mana of one color' answer N times (Gilded Lotus, #742)", () => {
    const opts = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, {
            produced: "{W3|U3|B3|R3|G3}",
            color_options: [["W", "U", "B", "R", "G"]],
          }),
        ],
      }),
    );
    expect(opts[4]).toMatchObject({ symbols: ["G", "G", "G"], caption: "3 Green", colors: ["G"] });
  });

  it("leaves a greyed ability one greyed option, so its reason is read once", () => {
    const opts = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, {
            produced: "{W|U|B|R|G}",
            exhausted: true,
            color_options: [["W", "U", "B", "R", "G"]],
          }),
        ],
      }),
    );
    expect(opts).toHaveLength(1);
    expect(opts[0].disabled).toBeTruthy();
    expect(opts[0].colors).toBeUndefined();
  });

  it("falls back to the server's prompt past the option cap", () => {
    const five = ["W", "U", "B", "R", "G"];
    expect(colorCombos([five, five, five])).toBeNull();
    const opts = manaAbilityOptions(
      card({
        mana_abilities: [
          ability(0, {
            produced: "{W|U|B|R|G}{W|U|B|R|G}{W|U|B|R|G}",
            color_options: [five, five, five],
          }),
        ],
      }),
    );
    expect(opts).toHaveLength(1);
    expect(opts[0].choice).toBe(true);
    expect(opts[0].colors).toBeUndefined();
  });

  it("sends color for one slot, colors for several, nothing for none", () => {
    expect(manaColorParams(undefined)).toEqual({});
    expect(manaColorParams([])).toEqual({});
    expect(manaColorParams(["R"])).toEqual({ color: "R" });
    expect(manaColorParams(["W", "U"])).toEqual({ colors: ["W", "U"] });
  });
});
