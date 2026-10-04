// abilityRow.test.ts — ADR 0117 §2: one predicate for "can this row be
// used". The click rule, the ability popover, the mana picker and the
// override menu all ask abilityRowBlocked, so each arm is pinned here
// once, in the ADR's order. abilityClick.render.test.ts checks that the
// click and the rendered popover agree over the same fixtures.

import { describe, expect, it } from "vitest";

import {
  ABILITY_EXHAUSTED,
  ABILITY_NOT_RIGHT_NOW,
  ACTIVATION_CONDITION_UNMET,
  ADDS_NO_MANA,
  EFFECT_STOPS_ABILITIES,
  NOT_ENOUGH_LIFE,
  abilityPopoverModel,
  abilityRowBlocked,
  abilityRowContext,
  battlefieldClickIntent,
  buildMenuSections,
  menuManaRows,
  type AbilityRow,
  type AbilityRowContext,
} from "./contextMenu.logic";
import { legalActionsOf } from "./legalActions";
import { manaAbilityOptions } from "./manaSource";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";

const ME = "me";

function zone(kind: string, owner: string | undefined, cards: CardView[] = []): ZoneView {
  return { kind, owner, count: cards.length, cards };
}
function seat(id: string, life = 40): PlayerView {
  return {
    id,
    name: id,
    seat: 0,
    life,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
  };
}
function gameView(cards: CardView[], extra: Partial<GameView> = {}): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME), seat("bob")],
    battlefield: zone("battlefield", undefined, cards),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
    ...extra,
  };
}

const permanent = (extra: Partial<CardView> = {}): CardView => ({
  instance_id: "c1",
  name: "Thing",
  owner: ME,
  controller: ME,
  type_line: "Artifact Creature — Golem",
  ...extra,
});

const ctxFor = (card: CardView, extra: Partial<AbilityRowContext> = {}): AbilityRowContext => ({
  ...abilityRowContext(card, { viewerID: ME }),
  ...extra,
});

const blocked = (
  a: AbilityRow,
  kind: "mana" | "activated",
  card: CardView = permanent(),
  extra: Partial<AbilityRowContext> = {},
) => abilityRowBlocked(a, kind, ctxFor(card, extra));

describe("abilityRowBlocked, one arm at a time (ADR 0117 §2)", () => {
  it("passes a row nothing stops", () => {
    expect(blocked({ tap_cost: true }, "activated")).toBe("");
    expect(blocked({ tap_cost: true }, "mana")).toBe("");
  });

  it("1. the card's restriction: cant_activate for activated rows, cant_activate_mana for mana rows", () => {
    const arrested = permanent({ restrictions: ["cant_activate"] });
    expect(blocked({}, "activated", arrested)).toBe(EFFECT_STOPS_ABILITIES);
    // Faith's Fetters spares mana abilities: the two bits are separate.
    expect(blocked({}, "mana", arrested)).toBe("");
    const fettered = permanent({ restrictions: ["cant_activate_mana"] });
    expect(blocked({}, "mana", fettered)).toBe(EFFECT_STOPS_ABILITIES);
    expect(blocked({}, "activated", fettered)).toBe("");
  });

  it("2. the row's own cant_activate clause (#1210), in the server's words", () => {
    const clause = "Activated abilities of creatures can't be activated";
    expect(blocked({ cant_activate: clause }, "activated")).toBe(clause);
    expect(blocked({ cant_activate: clause }, "mana")).toBe(clause);
  });

  it("3. a {T} cost on a tapped card — and only a {T} cost (CR 106.12)", () => {
    const tapped = permanent({ tapped: true });
    expect(blocked({ tap_cost: true }, "mana", tapped)).toBe("already tapped");
    // Vivi Ornitier's "{0}" ability: no {T}, so a tapped Vivi can use it.
    expect(blocked({ mana_cost: "{0}" }, "mana", tapped)).toBe("");
  });

  it("4. a {T} cost on a summoning-sick card (CR 302.6)", () => {
    const sick = permanent({ summoning_sick: true });
    expect(blocked({ tap_cost: true }, "mana", sick)).toBe("summoning sickness");
    expect(blocked({ mana_cost: "{1}" }, "activated", sick)).toBe("");
  });

  it("5. the shared costs: life, sacrifice, and the rest", () => {
    expect(blocked({ life_cost: 2 }, "mana", permanent(), { payerLife: 1 })).toBe(NOT_ENOUGH_LIFE);
    expect(blocked({ life_cost: 2 }, "mana", permanent(), { payerLife: 2 })).toBe("");
    expect(
      blocked(
        { sacrifice_options: { cards: [], min: 1, max: 1 }, sacrifice_label: "an artifact" },
        "activated",
      ),
    ).toMatch(/artifact/);
  });

  it("5. a planeswalker's rows, with the loyalty context: once per turn and an unpayable −N", () => {
    const walker = permanent({
      type_line: "Legendary Planeswalker — Teferi",
      counters: { loyalty: 2 },
    });
    expect(blocked({ loyalty_cost: 1 }, "activated", walker)).toBe("");
    expect(blocked({ loyalty_cost: -3 }, "activated", walker)).toMatch(/not enough loyalty/);
    const used = { ...walker, loyalty_activated: true };
    expect(blocked({ loyalty_cost: 1 }, "activated", used)).toBe("Already activated this turn");
  });

  it("6. timing_closed, in the panel's words when it passes them", () => {
    expect(
      blocked({ timing_closed: true }, "activated", permanent(), { timingWords: "Not your turn" }),
    ).toBe("Not your turn");
    expect(blocked({ timing_closed: true }, "activated", permanent(), { timingWords: "" })).toBe(
      ABILITY_NOT_RIGHT_NOW,
    );
  });

  it("7. exhausted, then condition_unmet (Vivi's turn and once-per-turn gates)", () => {
    expect(blocked({ exhausted: true, condition_unmet: true }, "mana")).toBe(ABILITY_EXHAUSTED);
    expect(blocked({ condition_unmet: true }, "mana")).toBe(ACTIVATION_CONDITION_UNMET);
  });

  it("8. adds_no_mana: the generic 'adds no mana right now' (ADR 0117 §5)", () => {
    expect(blocked({ adds_no_mana: true }, "mana")).toBe(ADDS_NO_MANA);
    expect(ADDS_NO_MANA).toBe("adds no mana right now");
  });

  it("9. the digest refuses a sorcery-speed row, or any row across the table, and nothing else", () => {
    const card = permanent();
    const view = gameView([card], { legal_actions: { pass: true, sources: {} } });
    const gate = legalActionsOf(view);
    expect(
      blocked({ sorcery_speed: true, ref: "own:0" }, "activated", card, { legalGate: gate }),
    ).toBe(ABILITY_NOT_RIGHT_NOW);
    // An instant-speed row the digest misses stays live (ADR 0105).
    expect(blocked({ ref: "own:0" }, "activated", card, { legalGate: gate })).toBe("");
    expect(blocked({ ref: "own:0" }, "activated", card, { legalGate: gate, across: true })).toBe(
      ABILITY_NOT_RIGHT_NOW,
    );
    // Mana rows never read the digest.
    expect(blocked({ sorcery_speed: true, ref: "own:0" }, "mana", card, { legalGate: gate })).toBe(
      "",
    );
  });

  it("keeps the order: the card's restriction outranks a tapped {T} cost", () => {
    const both = permanent({ tapped: true, restrictions: ["cant_activate"] });
    expect(blocked({ tap_cost: true }, "activated", both)).toBe(EFFECT_STOPS_ABILITIES);
  });
});

describe("the four callers agree", () => {
  const relicDraw = {
    index: 0,
    ref: "own:0",
    label: "{3}, {T}: Draw two cards, then discard a card.",
    tap_cost: true,
    mana_cost: "{3}",
  };
  const relic = (extra: Partial<CardView> = {}): CardView =>
    permanent({
      instance_id: "relic",
      name: "Relic of Sauron",
      type_line: "Legendary Artifact",
      mana_abilities: [
        {
          index: 0,
          ref: "own:m0",
          tap_cost: true,
          produced: "{U|B|R}{U|B|R}",
          color_options: [
            ["U", "B", "R"],
            ["U", "B", "R"],
          ],
        },
      ],
      activated_abilities: [relicDraw],
      ...extra,
    });

  it("the override menu greys what the click rule ignores, with the same reason", () => {
    const summoned = permanent({
      summoning_sick: true,
      mana_abilities: [{ index: 0, tap_cost: true, produced: "{G}" }],
    });
    const view = gameView([summoned]);
    const sections = buildMenuSections(view, summoned, ME, false);
    const row = sections.find((s) => s.id === "abilities")?.items[0];
    expect(row?.disabled).toBe(true);
    expect(row?.hint).toBe("summoning sickness");
    expect(manaAbilityOptions(summoned)[0].disabled).toBe("summoning sickness");
    expect(battlefieldClickIntent(summoned, ME, false, { manaClick: true, view })).toBe("none");
  });

  it("a row's own cant_activate reaches the override menu too", () => {
    const totem = relic({
      activated_abilities: [{ ...relicDraw, cant_activate: "Cursed Totem says no" }],
    });
    const view = gameView([totem]);
    const row = buildMenuSections(view, totem, ME, false)
      .find((s) => s.id === "abilities")
      ?.items.find((i) => i.id === "ability-0");
    expect(row?.disabled).toBe(true);
    expect(row?.hint).toBe("Cursed Totem says no");
  });

  it("Relic of Sauron: the popover, mana row first; the mana path when the draw row is refused; nothing tapped", () => {
    const opts = { manaClick: true, special: true, view: gameView([relic()]) };
    expect(battlefieldClickIntent(relic(), ME, false, opts)).toBe("popover");
    const model = abilityPopoverModel({
      card: relic(),
      viewerID: ME,
      view: opts.view,
      mana: true,
      activated: true,
      special: true,
    });
    expect(model.mana.map((r) => r.blocked)).toEqual([""]);
    expect(model.activated.map((r) => r.blocked)).toEqual([""]);

    const refused = relic({
      activated_abilities: [{ ...relicDraw, cant_activate: "an effect says no" }],
    });
    expect(battlefieldClickIntent(refused, ME, false, opts)).toBe("mana");

    expect(battlefieldClickIntent(relic({ tapped: true }), ME, false, opts)).toBe("none");
  });

  it("an opponent's Aura drawn on the viewer's creature lists none of its mana rows", () => {
    const aura = permanent({
      controller: "bob",
      owner: "bob",
      mana_abilities: [{ index: 0, tap_cost: true, produced: "{G}" }],
    });
    expect(menuManaRows(aura, ME)).toEqual([]);
    expect(menuManaRows(aura, undefined)).toHaveLength(1);
    expect(battlefieldClickIntent(aura, ME, false, { manaClick: true })).toBe("none");
  });
});
