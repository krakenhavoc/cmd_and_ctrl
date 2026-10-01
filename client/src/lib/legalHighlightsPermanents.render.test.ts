// @vitest-environment jsdom
//
// legalHighlightsPermanents.render.test.ts — ADR 0105 sub-PR 3 (#1789).
// What the board draws on PERMANENTS and in their menus from
// legalActions.ts, which only markup can show:
//
//   - a permanent with a live activated ability wears a bolt pip (with
//     a count from two up) and the ready ring;
//   - a permanent with a live mana ability worth marking (§4: Vivi's
//     free one) wears a drop pip and no ring, and a basic land wears
//     nothing;
//   - the ability popover puts the ready rows first, with the accent,
//     and greys a sorcery-speed row the exact digest leaves out. That
//     replaces the popover's old client-side sorcery-speed gate;
//   - highlights off take the pips, the ring and the accent away and
//     leave the gate where it was; no digest lights nothing and greys
//     nothing new; an opponent's permanent wears nothing;
//   - the card menu's ability rows take the same accent and order.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import ManaAbilityMenu from "./components/board/ManaAbilityMenu.svelte";
import { buildMenuSections } from "./contextMenu.logic";
import {
  NO_LEGAL_ACTIONS,
  legalActionsOf,
  visibleHighlights,
  type LegalActions,
} from "./legalActions";
import type {
  ActivatedAbilityView,
  CardView,
  GameView,
  LegalActionsView,
  ManaAbilityView,
  PlayerView,
} from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  localStorage.clear();
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

const ability = (index: number, extra: Partial<ActivatedAbilityView>): ActivatedAbilityView => ({
  index,
  ref: `own:${index}`,
  label: `ability ${index}`,
  ...extra,
});
const mana = (index: number, extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index,
  ref: `own:${index}`,
  ...extra,
});

// A Shikari-shaped equipment holder: an "only as a sorcery" equip row
// and an instant-speed pump row.
const blade = () =>
  permanent("blade", "Rogue's Blade", "Artifact — Equipment", {
    activated_abilities: [
      ability(0, { label: "Equip {2}", sorcery_speed: true }),
      ability(1, { label: "{1}: +1/+0" }),
    ],
  });
// Two live rows: the count.
const engine = () =>
  permanent("engine", "Twin Engine", "Artifact", {
    activated_abilities: [ability(0, { label: "{1}: Scry 1" }), ability(1, { label: "{2}: Draw" })],
  });
// #1621: a creature with a free, once-per-turn mana ability.
const vivi = () =>
  permanent("vivi", "Vivi Ornitier", "Legendary Creature — Wizard", {
    power: 0,
    toughness: 3,
    mana_abilities: [mana(0, { label: "Add X mana in any combination of {U} and/or {R}" })],
  });
const mountain = () =>
  permanent("mountain", "Mountain", "Basic Land — Mountain", {
    mana_abilities: [mana(0, { tap_cost: true, produced: "{R}", label: "Add {R}" })],
  });

// The digest the server would send on the viewer's main phase: the
// blade's pump row is live and its equip row is not (it can't pay
// it), both engine rows are live, Vivi's mana ability is live, and the
// Mountain can tap.
const digest: LegalActionsView = {
  pass: true,
  sources: {
    blade: { kinds: ["activate"], moves: 1, abilities: ["own:1"] },
    engine: { kinds: ["activate"], moves: 2, abilities: ["own:0", "own:1"] },
    vivi: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
    mountain: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
  },
};

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

const gameView = (cards: CardView[], extra: Partial<GameView> = {}): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat()],
    battlefield: zone("battlefield", undefined, cards),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    ...extra,
  }) as unknown as GameView;

const board = () => [blade(), engine(), vivi(), mountain()];

// mountPanel mounts the viewer's own panel the way Board does: `legal`
// is what may be highlighted, `legalGate` the frame's full lookup.
function mountPanel(view: GameView, legal: LegalActions, legalGate: LegalActions, isSelf = true) {
  const cards = view.battlefield.cards;
  const r = render(
    PlayerPanel as never,
    {
      seat: view.seats[0],
      isSelf,
      isActive: true,
      hasPriority: true,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: cards,
      exile: zone("exile", undefined),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: () => {},
      onManaAbilityCost: () => {},
      legal,
      legalGate,
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  const pip = (id: string, kind: "bolt" | "drop") =>
    tile(id).querySelector<HTMLElement>(`.ready-pip[data-pip="${kind}"]`);
  // Right-click opens the ability popover (admin overrides off).
  const openMenu = (id: string) => {
    tile(id).dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    flushSync();
    return [
      ...tile(id).querySelectorAll<HTMLButtonElement>(".mana-menu .menu-item:not([data-raw-tap])"),
    ];
  };
  return { ...r, tile, pip, openMenu };
}

const live = (view: GameView) => {
  const all = legalActionsOf(view);
  return mountPanel(view, visibleHighlights(all, true), all);
};

describe("permanents: the bolt and the drop", () => {
  const view = gameView(board(), { legal_actions: digest });

  it("a permanent with a live ability wears the bolt and the ring", () => {
    const p = live(view);
    expect(p.pip("blade", "bolt")).not.toBeNull();
    expect(p.tile("blade").classList.contains("ready")).toBe(true);
    // One live row: no count.
    expect(p.pip("blade", "bolt")?.querySelector(".pip-count")).toBeNull();
  });

  it("two live rows: the bolt says 2", () => {
    const p = live(view);
    expect(p.pip("engine", "bolt")?.querySelector(".pip-count")?.textContent).toBe("2");
  });

  it("Vivi's free mana ability: a drop pip and no ring", () => {
    const p = live(view);
    expect(p.pip("vivi", "drop")).not.toBeNull();
    expect(p.pip("vivi", "bolt")).toBeNull();
    expect(p.tile("vivi").classList.contains("ready")).toBe(false);
  });

  it("a basic land that can tap wears nothing", () => {
    const p = live(view);
    expect(p.tile("mountain").querySelector(".ready-pip")).toBeNull();
    expect(p.tile("mountain").classList.contains("ready")).toBe(false);
  });

  it("highlights off: no pip, no ring", () => {
    const all = legalActionsOf(view);
    const p = mountPanel(view, visibleHighlights(all, false), all);
    expect(p.container.querySelector(".ready-pip")).toBeNull();
    expect(p.container.querySelector(".card.ready")).toBeNull();
  });

  it("no digest and no list: nothing lights", () => {
    const quiet = gameView(board());
    const all = legalActionsOf(quiet);
    const p = mountPanel(quiet, visibleHighlights(all, true), all);
    expect(p.container.querySelector(".ready-pip")).toBeNull();
    expect(p.container.querySelector(".card.ready")).toBeNull();
  });

  it("an opponent's panel wears nothing, whatever the lookup holds", () => {
    const all = legalActionsOf(view);
    const p = mountPanel(view, all, all, false);
    expect(p.container.querySelector(".ready-pip")).toBeNull();
    expect(p.container.querySelector(".card.ready")).toBeNull();
  });
});

describe("the ability popover: ready rows first, and the digest's gate", () => {
  it("puts the live row first with the accent, and greys the equip the digest leaves out", () => {
    const p = live(gameView(board(), { legal_actions: digest }));
    const rows = p.openMenu("blade");
    expect(rows.map((b) => b.querySelector(".label")?.textContent)).toEqual([
      "{1}: +1/+0",
      "Equip {2}",
    ]);
    expect(rows[0].classList.contains("ready")).toBe(true);
    expect(rows[0].disabled).toBe(false);
    expect(rows[1].classList.contains("ready")).toBe(false);
    expect(rows[1].disabled).toBe(true);
    expect(rows[1].title).toBe("Can't activate this right now");
  });

  it("highlights off: the order and the accent go, the gate stays", () => {
    const view = gameView(board(), { legal_actions: digest });
    const all = legalActionsOf(view);
    const p = mountPanel(view, visibleHighlights(all, false), all);
    const rows = p.openMenu("blade");
    expect(rows.map((b) => b.querySelector(".label")?.textContent)).toEqual([
      "Equip {2}",
      "{1}: +1/+0",
    ]);
    expect(rows.some((b) => b.classList.contains("ready"))).toBe(false);
    expect(rows[0].disabled).toBe(true);
    expect(rows[1].disabled).toBe(false);
  });

  it("no digest: no accent, and no row greyed that the row fields do not grey", () => {
    const quiet = gameView(board());
    const all = legalActionsOf(quiet);
    const p = mountPanel(quiet, visibleHighlights(all, true), all);
    const rows = p.openMenu("blade");
    expect(rows.some((b) => b.classList.contains("ready"))).toBe(false);
    expect(rows.every((b) => !b.disabled)).toBe(true);
  });

  it("the capped list is not exact: it lights rows, but greys none", () => {
    const capped = gameView(board(), {
      legal_moves: [
        {
          type: "activate_ability",
          player: ME,
          kind: "activate",
          label: "pump",
          source: "blade",
          params: { card_id: "blade", ref: "own:1" },
        },
      ],
    });
    const p = live(capped);
    const rows = p.openMenu("blade");
    expect(rows[0].classList.contains("ready")).toBe(true);
    expect(rows.every((b) => !b.disabled)).toBe(true);
  });

  it("Vivi's mana row is ready in her popover", () => {
    const p = live(gameView(board(), { legal_actions: digest }));
    const rows = p.openMenu("vivi");
    expect(rows).toHaveLength(1);
    expect(rows[0].classList.contains("ready")).toBe(true);
  });
});

describe("ManaAbilityMenu: timing is the server's, the words are the panel's", () => {
  const exact = (abilities: string[]) =>
    legalActionsOf(
      gameView([], {
        legal_actions: { pass: true, sources: { c: { kinds: ["activate"], moves: 1, abilities } } },
      }),
    );

  it("a sorcery-speed row the digest lists stays live, whatever the panel's words say", () => {
    // Leonin Shikari's equip in combat: the printed clause says
    // sorcery, a per-player statement opens it, and the old client gate
    // greyed it anyway.
    const { container } = render(ManaAbilityMenu, {
      abilities: [],
      tapped: false,
      onActivate: () => {},
      activated: [ability(0, { label: "Equip {1}", sorcery_speed: true })],
      timingReason: "Sorcery-speed only",
      cardID: "c",
      legalGate: exact(["own:0"]),
    });
    expect(container.querySelector<HTMLButtonElement>(".menu-item")!.disabled).toBe(false);
  });

  it("a row the server shut (timing_closed) greys with the panel's words", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [],
      tapped: false,
      onActivate: () => {},
      activated: [ability(0, { label: "+1: Draw", loyalty_cost: 1, timing_closed: true })],
      timingReason: "Not your turn",
      cardID: "c",
      legalGate: NO_LEGAL_ACTIONS,
    });
    const row = container.querySelector<HTMLButtonElement>(".menu-item")!;
    expect(row.disabled).toBe(true);
    expect(row.title).toBe("Not your turn");
  });

  it("no digest: a sorcery-speed row is not greyed on the panel's words alone", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [],
      tapped: false,
      onActivate: () => {},
      activated: [ability(0, { label: "Equip {1}", sorcery_speed: true })],
      timingReason: "Not your priority",
    });
    expect(container.querySelector<HTMLButtonElement>(".menu-item")!.disabled).toBe(false);
  });

  it("an instant-speed row the digest leaves out is not greyed", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [],
      tapped: false,
      onActivate: () => {},
      activated: [ability(0, { label: "{1}: +1/+0" })],
      cardID: "c",
      legalGate: exact([]),
    });
    expect(container.querySelector<HTMLButtonElement>(".menu-item")!.disabled).toBe(false);
  });
});

describe("the card menu: ready ability rows first, with the flag", () => {
  const view = gameView(board(), { legal_actions: digest });
  const abilityRows = (legal: LegalActions) =>
    buildMenuSections(view, blade(), ME, false, legal).find((s) => s.id === "abilities")!.items;

  it("marks and lifts the row the digest lists", () => {
    const rows = abilityRows(legalActionsOf(view));
    expect(rows.map((i) => i.label)).toEqual(["{1}: +1/+0", "Equip {2}"]);
    expect(rows.map((i) => i.ready === true)).toEqual([true, false]);
  });

  it("highlights off: the menu is exactly what it was", () => {
    const rows = abilityRows(NO_LEGAL_ACTIONS);
    expect(rows.map((i) => i.label)).toEqual(["Equip {2}", "{1}: +1/+0"]);
    expect(rows.some((i) => i.ready)).toBe(false);
  });
});
