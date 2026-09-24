// @vitest-environment jsdom
//
// #1443 — pick the colour BEFORE tapping. #1438 made a left-click on a
// mana source tap it for mana, but a source whose output is a choice of
// colours still asked the colour in a second step, after the tap: a
// painland was two picks, and Birds of Paradise / Command Tower asked in
// the centred mana_pick modal, which cannot be cancelled because the
// source is already tapped.
//
// The server now publishes each ability's `color_options` (the list the
// mana_pick would carry) and takes the answer up front as `color`. These
// tests mount the REAL Board — the click lands in PlayerPanel, the
// anchored picker is Board's, and the pick goes out through Board's
// handleMenuActivate — and check that every source is one pick, that the
// colour rides the one action, and that Escape sends nothing at all.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { get } from "svelte/store";

import Board from "./components/board/Board.svelte";
import { closeManaSourcePicker, manaSourcePicker } from "./manaSourcePicker";
import type { ActionType, CardView, GameView, ManaAbilityView, PlayerView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

// jsdom ships no ResizeObserver / IntersectionObserver and the board
// measures itself with both (boardFreeze.render.test.ts stubs them the
// same way). Stubs, not mocks: nothing here is under test.
class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
});

afterEach(() => {
  closeManaSourcePicker();
  cleanup();
});

const ME = "me";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const ability = (index: number, extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index,
  tap_cost: true,
  ...extra,
});

const permanent = (id: string, name: string, typeLine: string, extra: Partial<CardView> = {}) =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: typeLine,
    ...extra,
  }) as CardView;

// Battlefield Forge as the server ships it after #1443: the painless
// {C} first, the pain dual second with its one picking slot.
const forge = () =>
  permanent("forge", "Battlefield Forge", "Land", {
    mana_abilities: [
      ability(0, { produced: "{C}", label: "Add {C}" }),
      ability(1, {
        produced: "{R|W}",
        label: "Add {R} or {W}. This land deals 1 damage to you.",
        color_options: [["R", "W"]],
      }),
    ],
  });

// Birds of Paradise under a mono-green commander: all five, identity
// first (#843) — the server's order, which the picker must keep.
const birds = () =>
  permanent("birds", "Birds of Paradise", "Creature — Bird", {
    power: 0,
    toughness: 1,
    mana_abilities: [
      ability(0, {
        produced: "{W|U|B|R|G}",
        label: "Add one mana of any color",
        color_options: [["G", "W", "U", "B", "R"]],
      }),
    ],
  });

// Command Tower narrowed to the identity (CR 903.4f).
const tower = (identity: string[]) =>
  permanent("tower", "Command Tower", "Land", {
    mana_abilities: [
      ability(0, {
        produced: "{W|U|B|R|G}",
        label: "Add one mana of any color in your commander's color identity",
        color_options: [identity],
      }),
    ],
  });

const swamp = () =>
  permanent("swamp", "Swamp", "Basic Land — Swamp", {
    mana_abilities: [ability(0, { produced: "{B}", label: "Add {B}" })],
  });

const seat = (): PlayerView =>
  ({
    id: ME,
    name: "Me",
    seat: 0,
    life: 40,
    library: zone("library", ME),
    hand: zone("hand", ME),
    graveyard: zone("graveyard", ME),
    command: zone("command", ME),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (battlefield: CardView[]): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat()],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountBoard(cards: CardView[]) {
  const sent: Sent[] = [];
  const r = render(
    Board as never,
    {
      view: gameView(cards),
      viewerID: ME,
      isAdmin: false,
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  const picker = () => document.querySelector<HTMLElement>(".mana-source-picker");
  const options = () => Array.from(picker()?.querySelectorAll<HTMLElement>(".mana-option") ?? []);
  const symbols = (el: Element) =>
    Array.from(el.querySelectorAll("svg[data-symbol]")).map((s) => s.getAttribute("data-symbol"));
  return { ...r, sent, tile, picker, options, symbols };
}

const key = (k: string) => {
  window.dispatchEvent(new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true }));
  flushSync();
};

describe("one pick per mana source (#1443)", () => {
  it("offers a painland's final results — {C}, {R} and {W} with its damage — and sends the colour", () => {
    const b = mountBoard([forge()]);
    click(b.tile("forge"));
    expect(b.sent).toEqual([]);
    const opts = b.options();
    expect(opts.map((o) => b.symbols(o))).toEqual([["C"], ["R"], ["W"]]);
    expect(opts[0].textContent).not.toContain("damage");
    expect(opts[1].textContent).toContain("deals 1 damage to you");
    expect(opts[2].textContent).toContain("deals 1 damage to you");
    // Nothing says "pick the color next" any more: this IS the pick.
    expect(b.picker()?.textContent).not.toMatch(/next/i);

    click(opts[2]);
    expect(b.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "forge", ability_index: 1, color: "W" },
      },
    ]);
    expect(get(manaSourcePicker)).toBeNull();
  });

  it("sends the painless {C} with no colour at all", () => {
    const b = mountBoard([forge()]);
    click(b.tile("forge"));
    click(b.options()[0]);
    expect(b.sent).toEqual([
      { type: "activate_mana_ability", params: { card_id: "forge", ability_index: 0 } },
    ]);
  });

  it("anchors Birds of Paradise's five colours at the card, in the server's order", () => {
    const b = mountBoard([birds()]);
    click(b.tile("birds"));
    expect(b.sent).toEqual([]);
    expect(b.options().map((o) => b.symbols(o))).toEqual([["G"], ["W"], ["U"], ["B"], ["R"]]);
    key("3");
    expect(b.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "birds", ability_index: 0, color: "U" },
      },
    ]);
  });

  it("offers Command Tower only its identity colours", () => {
    const b = mountBoard([tower(["U", "R"])]);
    click(b.tile("tower"));
    expect(b.options().map((o) => b.symbols(o))).toEqual([["U"], ["R"]]);
    click(b.options()[1]);
    expect(b.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "tower", ability_index: 0, color: "R" },
      },
    ]);
  });

  it("taps a mono-identity Command Tower at once, naming its one colour", () => {
    const b = mountBoard([tower(["G"])]);
    click(b.tile("tower"));
    expect(b.picker()).toBeNull();
    expect(b.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "tower", ability_index: 0, color: "G" },
      },
    ]);
  });

  it("still taps a fixed single-output source at once, with no colour", () => {
    const b = mountBoard([swamp()]);
    click(b.tile("swamp"));
    expect(b.picker()).toBeNull();
    expect(b.sent).toEqual([
      { type: "activate_mana_ability", params: { card_id: "swamp", ability_index: 0 } },
    ]);
  });

  for (const [name, card] of [
    ["painland", forge],
    ["Birds of Paradise", birds],
    ["Command Tower", () => tower(["U", "R"])],
  ] as const) {
    it(`Escape on the ${name} picker sends nothing, so nothing is tapped`, () => {
      const b = mountBoard([card()]);
      click(b.tile(card().instance_id));
      expect(b.picker()).not.toBeNull();
      key("Escape");
      expect(b.picker()).toBeNull();
      expect(get(manaSourcePicker)).toBeNull();
      expect(b.sent).toEqual([]);
    });
  }
});
