// @vitest-environment jsdom
//
// anyPlayerActivation.render.test.ts — ADR 0106 §1 (#1793). What the
// board draws and does on ANOTHER player's permanent that carries an
// "Any player may activate this ability" row (CR 602.2), which only
// markup can show:
//
//   - it wears the bolt pip and the ready ring exactly when the
//     viewer's own digest lists it (decision 6), on a frame where
//     highlights are live;
//   - a left-click on it opens the card menu, and that menu's only row
//     is the any-player row, which hands the activation to Board;
//   - its right-click popover lists only that row, greyed when the
//     exact digest leaves it out.

import { describe, it, expect, afterEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import CardContextMenu from "./components/board/CardContextMenu.svelte";
import { cardMenu, closeCardMenu } from "./contextMenu";
import { legalActionsOf, visibleHighlights, type LegalActions } from "./legalActions";
import type {
  ActivatedAbilityView,
  CardView,
  GameView,
  LegalActionsView,
  PlayerView,
} from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  closeCardMenu();
  localStorage.clear();
});

const ME = "me";
const ALICE = "alice";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const row = (index: number, extra: Partial<ActivatedAbilityView>): ActivatedAbilityView => ({
  index,
  ref: `own:${index}`,
  label: `ability ${index}`,
  ...extra,
});

const xantcha = (): CardView =>
  ({
    instance_id: "xantcha",
    name: "Xantcha, Sleeper Agent",
    owner: ALICE,
    controller: ALICE,
    known_by_you: true,
    type_line: "Legendary Creature — Phyrexian Minion",
    power: 5,
    toughness: 5,
    activated_abilities: [
      row(0, { label: "{3}: Draw a card; its controller loses 2 life", any_player: true }),
      row(1, { label: "{1}: Alice only" }),
    ],
  }) as CardView;

const seat = (id: string, idx: number): PlayerView =>
  ({
    id,
    name: id === ME ? "Me" : "Alice",
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

// The viewer's main phase, holding priority.
const listed: LegalActionsView = {
  pass: true,
  sources: { xantcha: { kinds: ["activate"], moves: 1, abilities: ["own:0"] } },
};

const gameView = (extra: Partial<GameView> = {}): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, 0), seat(ALICE, 1)],
    battlefield: zone("battlefield", undefined, [xantcha()]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    ...extra,
  }) as unknown as GameView;

// mountAlice mounts ALICE's panel as the viewer ME sees it, the way
// Board does: `legal` is what may be highlighted, `legalGate` the
// frame's full lookup, and the activation handler is wired.
function mountAlice(view: GameView, legal: LegalActions, legalGate: LegalActions) {
  const activated: [string, number][] = [];
  const r = render(
    PlayerPanel as never,
    {
      seat: view.seats[1],
      isSelf: false,
      isActive: false,
      hasPriority: false,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: view.battlefield.cards,
      exile: zone("exile", undefined),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {
        throw new Error("a non-controller's click must not tap Alice's permanent");
      },
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: (card: CardView, index: number) =>
        activated.push([card.instance_id, index]),
      onManaAbilityCost: () => {},
      legal,
      legalGate,
    } as never,
  );
  const tile = () => r.container.querySelector<HTMLElement>(`.card[data-instance-id="xantcha"]`)!;
  const bolt = () => tile().querySelector<HTMLElement>(`.ready-pip[data-pip="bolt"]`);
  const openPopover = () => {
    tile().dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    flushSync();
    return [
      ...tile().querySelectorAll<HTMLButtonElement>(".mana-menu .menu-item:not([data-raw-tap])"),
    ];
  };
  return { ...r, tile, bolt, openPopover, activated };
}

const live = (view: GameView) => {
  const all = legalActionsOf(view);
  return mountAlice(view, visibleHighlights(all, true), all);
};

describe("the bolt pip on another player's permanent", () => {
  it("shows, with the ring, when the viewer's digest lists the permanent", () => {
    const p = live(gameView({ legal_actions: listed }));
    expect(p.bolt()).not.toBeNull();
    expect(p.tile().classList.contains("ready")).toBe(true);
  });

  it("does not show when the digest has no entry for it", () => {
    const p = live(gameView({ legal_actions: { pass: true, sources: {} } }));
    expect(p.bolt()).toBeNull();
    expect(p.tile().classList.contains("ready")).toBe(false);
  });

  it("does not show while highlights are not live (autopass passing, or the setting off)", () => {
    const view = gameView({ legal_actions: listed });
    const all = legalActionsOf(view);
    const p = mountAlice(view, visibleHighlights(all, false), all);
    expect(p.bolt()).toBeNull();
  });
});

describe("the left-click on another player's permanent", () => {
  it("opens the card menu rather than tapping it", () => {
    const p = live(gameView({ legal_actions: listed }));
    p.tile().click();
    flushSync();
    expect(get(cardMenu)?.card.instance_id).toBe("xantcha");
  });
});

describe("the card menu on another player's permanent", () => {
  it("shows only the any-player row, and choosing it hands the activation to Board", () => {
    const view = gameView({ legal_actions: listed });
    const all = legalActionsOf(view);
    const onActivate = vi.fn();
    const onClose = vi.fn();
    const r = render(
      CardContextMenu as never,
      {
        view,
        viewerID: ME,
        isAdmin: false,
        open: { card: view.battlefield.cards[0], x: 10, y: 10 },
        sendAction: () => {},
        onActivate,
        onClose,
        legal: all,
        legalGate: all,
      } as never,
    );
    const rows = [...r.container.querySelectorAll<HTMLButtonElement>(".ctx-item")];
    expect(rows.map((b) => b.querySelector(".ctx-text")?.textContent)).toEqual([
      "{3}: Draw a card; its controller loses 2 life",
    ]);
    expect(rows[0].classList.contains("ready")).toBe(true);
    rows[0].click();
    flushSync();
    expect(onActivate).toHaveBeenCalledWith(expect.objectContaining({ instance_id: "xantcha" }), {
      kind: "ability",
      index: 0,
    });
    expect(onClose).toHaveBeenCalled();
  });
});

describe("the popover on another player's permanent", () => {
  it("lists only the any-player row, and activates it for the viewer", () => {
    const p = live(gameView({ legal_actions: listed }));
    const rows = p.openPopover();
    expect(rows.map((b) => b.querySelector(".label")?.textContent)).toEqual([
      "{3}: Draw a card; its controller loses 2 life",
    ]);
    expect(rows[0].disabled).toBe(false);
    rows[0].click();
    flushSync();
    expect(p.activated).toEqual([["xantcha", 0]]);
  });

  it("greys the row when the exact digest leaves it out", () => {
    const p = live(gameView({ legal_actions: { pass: true, sources: {} } }));
    const rows = p.openPopover();
    expect(rows).toHaveLength(1);
    expect(rows[0].disabled).toBe(true);
  });
});
