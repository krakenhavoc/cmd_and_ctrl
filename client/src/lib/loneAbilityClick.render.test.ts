// @vitest-environment jsdom
//
// loneAbilityClick.render.test.ts — ADR 0117 §1 as amended on
// 2026-10-04 (#2201), on the real Board. A permanent whose ONE usable
// row is not a mana row activates it on a left-click, through the very
// path its popover row takes (Board's activation flow: costs, X, modes,
// targets), and opens no popover. Two or more usable rows still open
// the popover. "Usable" is ADR 0117 §2's predicate, so a greyed second
// row does not count.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import { abilityPopover } from "./abilityPopover";
import { closeManaSourcePicker, manaSourcePicker } from "./manaSourcePicker";
import { targeting } from "./targeting";
import type {
  ActionType,
  ActivatedAbilityView,
  CardView,
  GameView,
  ManaAbilityView,
  PlayerView,
} from "./protocol";
import { subscribe, type TutorialEvent } from "./tutorialBus";
import { render, click, cleanup } from "./test/render.svelte";

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
  targeting.set(null);
  cleanup();
});

const ME = "me";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
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

// The viewer's precombat main phase, holding priority, the stack empty.
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
  params?: Record<string, unknown>;
}

function mountBoard(cards: CardView[]) {
  const sent: Sent[] = [];
  const r = render(
    Board as never,
    {
      view: gameView(cards),
      viewerID: ME,
      isAdmin: false,
      sendAction: (type: ActionType, params?: Record<string, unknown>) =>
        sent.push({ type, params }),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  const activations = () => sent.filter((s) => s.type === "activate_ability");
  return { ...r, sent, tile, activations };
}

// Polluted Delta as the server ships it: one activated row, every cost
// component on it, no mana ability.
const fetchRow: ActivatedAbilityView = {
  index: 0,
  ref: "own:0",
  label:
    "{T}, Pay 1 life, Sacrifice this land: Search your library for an Island or Swamp card, put it onto the battlefield, then shuffle.",
  tap_cost: true,
  life_cost: 1,
  sacrifice_self: true,
};
const fetchLand = (extra: Partial<CardView> = {}) =>
  permanent("delta", "Polluted Delta", "Land", { activated_abilities: [fetchRow], ...extra });

// Urborg, Tomb of Yawgmoth makes every land a Swamp: the fetch land
// gains the Swamp's intrinsic "{T}: Add {B}." beside its own ability
// (CR 305.6).
const urborgSwamp: ManaAbilityView = {
  index: 0,
  ref: "land:B",
  tap_cost: true,
  produced: "{B}",
  label: "{T}: Add {B}.",
};

const prodigal = () =>
  permanent("prodigal", "Prodigal Sorcerer", "Creature — Human Wizard", {
    power: 1,
    toughness: 1,
    activated_abilities: [
      {
        index: 0,
        ref: "own:0",
        label: "{T}: This creature deals 1 damage to any target.",
        tap_cost: true,
        legal_targets: { players: [ME], cards: ["prodigal"], min: 1, max: 1 },
      },
    ],
  });

const bear = () =>
  permanent("bear", "Grizzly Bears", "Creature — Bear", { power: 2, toughness: 2 });

const bonesplitter = () =>
  permanent("sword", "Bonesplitter", "Artifact — Equipment", {
    activated_abilities: [
      {
        index: 0,
        ref: "own:0",
        label: "Equip {1}",
        mana_cost: "{1}",
        sorcery_speed: true,
        legal_targets: { cards: ["bear"], min: 1, max: 1 },
      },
    ],
  });

// A catalogued planeswalker at two loyalty: its −3 cannot be paid
// (CR 606.6), so its +1 is the one usable row.
const walker = (loyalty: number) =>
  permanent("walker", "Teferi, Time Raveler", "Legendary Planeswalker — Teferi", {
    counters: { loyalty },
    activated_abilities: [
      { index: 0, ref: "own:0", label: "+1: …", loyalty_cost: 1, sorcery_speed: true },
      { index: 1, ref: "own:1", label: "−3: …", loyalty_cost: -3, sorcery_speed: true },
    ],
  });

// A creature cast this turn: its {T} row is greyed by summoning
// sickness (CR 302.6); its pump has no {T} and is usable.
const sickPinger = () =>
  permanent("pinger", "Firebreathing Pinger", "Creature — Goblin", {
    power: 1,
    toughness: 1,
    summoning_sick: true,
    activated_abilities: [
      {
        index: 0,
        ref: "own:0",
        label: "{T}: This creature deals 1 damage to target player.",
        tap_cost: true,
        legal_targets: { players: [ME], min: 1, max: 1 },
      },
      { index: 1, ref: "own:1", label: "{R}: This creature gets +1/+0.", mana_cost: "{R}" },
    ],
  });

const relic = () =>
  permanent("relic", "Relic of Sauron", "Legendary Artifact", {
    mana_abilities: [
      {
        index: 0,
        ref: "own:m0",
        tap_cost: true,
        label: "{T}: Add two mana in any combination of {U}, {B}, and/or {R}.",
        produced: "{U|B|R}{U|B|R}",
        color_options: [
          ["U", "B", "R"],
          ["U", "B", "R"],
        ],
      },
    ],
    activated_abilities: [
      {
        index: 0,
        ref: "own:0",
        label: "{3}, {T}: Draw two cards, then discard a card.",
        tap_cost: true,
        mana_cost: "{3}",
      },
    ],
  });

describe("one usable ability activates on click (#2201)", () => {
  it("a fetch land fetches in one click: the activation goes out, no popover opens", () => {
    const seen: TutorialEvent[] = [];
    const off = subscribe((e) => seen.push(e));
    try {
      const b = mountBoard([fetchLand()]);
      click(b.tile("delta"));
      expect(b.activations()).toHaveLength(1);
      expect(b.activations()[0].params).toMatchObject({
        source_card_id: "delta",
        ability_index: 0,
      });
      expect(get(abilityPopover)).toBeNull();
      expect(get(manaSourcePicker)).toBeNull();
      // No raw tap rides along, and no menu means nothing for the
      // tutorial's step 7.
      expect(b.sent.some((s) => s.type === "tap")).toBe(false);
      expect(seen).not.toContain("ability-menu-opened");
    } finally {
      off();
    }
  });

  it("the same fetch land under Urborg opens the popover: two usable rows", () => {
    const b = mountBoard([fetchLand({ mana_abilities: [urborgSwamp] })]);
    click(b.tile("delta"));
    expect(get(abilityPopover)?.cardID).toBe("delta");
    expect(b.sent).toEqual([]);
  });

  it("Prodigal Sorcerer starts its targeting", () => {
    const b = mountBoard([prodigal()]);
    click(b.tile("prodigal"));
    const t = get(targeting);
    expect(t?.card.instance_id).toBe("prodigal");
    expect(t?.ability?.index).toBe(0);
    expect(get(abilityPopover)).toBeNull();
    // The activation waits for the target.
    expect(b.activations()).toEqual([]);
  });

  it("an Equipment with Equip usable starts the equip flow: the creature choice", () => {
    const b = mountBoard([bonesplitter(), bear()]);
    click(b.tile("sword"));
    const t = get(targeting);
    expect(t?.card.instance_id).toBe("sword");
    expect(t?.ability?.index).toBe(0);
    expect([...(t?.legal?.cards ?? [])]).toEqual(["bear"]);
    expect(get(abilityPopover)).toBeNull();
  });

  it("a planeswalker with exactly one usable loyalty ability activates it", () => {
    const b = mountBoard([walker(2)]);
    click(b.tile("walker"));
    expect(b.activations()).toHaveLength(1);
    expect(b.activations()[0].params).toMatchObject({
      source_card_id: "walker",
      ability_index: 0,
    });
    expect(get(abilityPopover)).toBeNull();
  });

  it("a planeswalker with two usable loyalty abilities opens the popover", () => {
    const b = mountBoard([walker(4)]);
    click(b.tile("walker"));
    expect(get(abilityPopover)?.cardID).toBe("walker");
    expect(b.sent).toEqual([]);
  });

  it("a summoning-sick creature activates its one usable row; the greyed {T} row does not count", () => {
    const b = mountBoard([sickPinger()]);
    click(b.tile("pinger"));
    expect(b.activations()).toHaveLength(1);
    expect(b.activations()[0].params).toMatchObject({
      source_card_id: "pinger",
      ability_index: 1,
    });
    expect(get(abilityPopover)).toBeNull();
    expect(get(targeting)).toBeNull();
  });

  it("Relic of Sauron with both rows usable opens the popover", () => {
    const b = mountBoard([relic()]);
    click(b.tile("relic"));
    expect(get(abilityPopover)?.cardID).toBe("relic");
    expect(get(manaSourcePicker)).toBeNull();
    expect(b.sent).toEqual([]);
  });
});
