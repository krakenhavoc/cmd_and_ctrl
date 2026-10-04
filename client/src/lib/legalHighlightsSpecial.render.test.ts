// @vitest-environment jsdom
//
// legalHighlightsSpecial.render.test.ts — ADR 0105 sub-PR 4 (#1789).
// Special actions in the ordinary popover, and the pips as the touch
// route into it, which only markup can show:
//
//   - a foretell-able hand card that cannot be cast wears the star and
//     the ring and is NOT dimmed; its popover has a foretell row that
//     sends the admin menu's own payload;
//   - a cyclable hand card wears the bolt; a hand card with nothing to
//     do is still dimmed;
//   - a face-down permanent that can be turned face up wears the star;
//   - a pip is a button: a click or Enter on it opens the popover and
//     never fires the card's own click;
//   - highlights off: no pip, no ring, no accent, but the rows and the
//     gates stay; no digest: no pip and no row newly greyed.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import Card from "./components/board/Card.svelte";
import { legalActionsOf, readyPips, visibleHighlights, type LegalActions } from "./legalActions";
import type {
  ActivatedAbilityView,
  CardView,
  GameView,
  LegalActionsView,
  LegalMoveView,
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

const mine = (id: string, name: string, typeLine: string, extra: Partial<CardView> = {}) =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: typeLine,
    ...extra,
  }) as CardView;

const ability = (index: number, label: string): ActivatedAbilityView => ({
  index,
  ref: `own:${index}`,
  label,
});

// In hand: a foretell card the seat cannot cast yet but can foretell,
// a cycler it can only cycle, and a card it can do nothing with.
const foretell = () =>
  mine("dwarf", "Dwarven Hammer", "Artifact — Equipment", {
    mana_cost: "{2}{R}",
    special_actions: [{ kind: "foretell", label: "Foretell {2}", cost: "{2}", available: true }],
  });
const cycler = () =>
  mine("cycler", "Shark Typhoon", "Enchantment", {
    mana_cost: "{5}{U}",
    zone_abilities: [ability(0, "Cycling {X}{1}{U}")],
  });
const dead = () => mine("dead", "Ugin", "Legendary Planeswalker — Ugin", { mana_cost: "{8}" });
// On the battlefield: the viewer's own face-down morph.
const morph = () =>
  mine("morph", "", "", {
    face_down: true,
    face_visible: true,
    face_down_kind: "morph",
    power: 2,
    toughness: 2,
    special_actions: [
      { kind: "turn_face_up", label: "Turn face up {1}{U}", cost: "{1}{U}", available: true },
    ],
  });

const digest: LegalActionsView = {
  pass: true,
  sources: {
    dwarf: { kinds: ["special_action"], moves: 1, special_actions: ["foretell"] },
    cycler: { kinds: ["activate"], moves: 1, abilities: ["own:0"] },
    morph: { kinds: ["special_action"], moves: 1, special_actions: ["turn_face_up"] },
  },
};
// The capped list the server ships beside it. No cast move for any hand
// card, so canCastFromHand says no and the hand's dim applies.
const moves: LegalMoveView[] = [
  { type: "pass_priority", player: ME, kind: "pass", label: "pass" },
  {
    type: "special_action",
    player: ME,
    kind: "special_action",
    label: "Foretell {2}",
    source: "dwarf",
    params: { card_id: "dwarf", kind: "foretell" },
  },
  {
    type: "activate_ability",
    player: ME,
    kind: "activate",
    label: "Cycling",
    source: "cycler",
    params: { card_id: "cycler", ref: "own:0" },
  },
  {
    type: "special_action",
    player: ME,
    kind: "special_action",
    label: "Turn face up {1}{U}",
    source: "morph",
    params: { card_id: "morph", kind: "turn_face_up" },
  },
];

const seat = (hand: CardView[]): PlayerView =>
  ({
    id: ME,
    name: "Me",
    seat: 0,
    life: 40,
    library: zone("library", ME),
    hand: zone("hand", ME, hand),
    graveyard: zone("graveyard", ME),
    command: zone("command", ME),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (extra: Partial<GameView> = {}): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat([foretell(), cycler(), dead()])],
    battlefield: zone("battlefield", undefined, [morph()]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    ...extra,
  }) as unknown as GameView;

function mountPanel(view: GameView, legal: LegalActions, legalGate: LegalActions) {
  const sendAction = vi.fn();
  const onPlayCard = vi.fn();
  const r = render(
    PlayerPanel as never,
    {
      seat: view.seats[0],
      isSelf: true,
      isActive: true,
      hasPriority: true,
      viewerID: ME,
      isAdmin: false,
      sendAction,
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
      onTapToggle: () => {},
      onPlayCard,
      onDrawCard: () => {},
      onActivateAbility: () => {},
      onManaAbilityCost: () => {},
      legal,
      legalGate,
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  const slot = (id: string) => tile(id).closest<HTMLElement>(".hand-slot")!;
  const pip = (id: string, kind: "star" | "bolt" | "drop") =>
    tile(id).querySelector<HTMLButtonElement>(`.ready-pip[data-pip="${kind}"]`);
  const menuRows = (id: string) => [
    // ADR 0117 §3: not the Sandbox rows — the raw tap, which every own
    // permanent has, and (ADR 0118 §2, strict being the default since
    // §1) Cast anyway, which every castable hand card has.
    ...tile(id).querySelectorAll<HTMLButtonElement>(
      ".mana-menu .menu-item:not([data-raw-tap]):not([data-cast-anyway])",
    ),
  ];
  const rightClick = (id: string) => {
    tile(id).dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    flushSync();
    return menuRows(id);
  };
  return { ...r, sendAction, onPlayCard, tile, slot, pip, menuRows, rightClick };
}

const live = (view: GameView) => {
  const all = legalActionsOf(view);
  return mountPanel(view, visibleHighlights(all, true), all);
};

const withDigest = () => gameView({ legal_actions: digest, legal_moves: moves });

describe("hand cards with only a special action or a hand ability", () => {
  it("a foretell-able card: the star, the ring, and no dim", () => {
    const p = live(withDigest());
    expect(p.pip("dwarf", "star")).not.toBeNull();
    expect(p.tile("dwarf").classList.contains("ready")).toBe(true);
    expect(p.slot("dwarf").classList.contains("timing-disabled")).toBe(false);
  });

  it("a cyclable card: the bolt, the ring, and no dim", () => {
    const p = live(withDigest());
    expect(p.pip("cycler", "bolt")).not.toBeNull();
    expect(p.pip("cycler", "star")).toBeNull();
    expect(p.tile("cycler").classList.contains("ready")).toBe(true);
    expect(p.slot("cycler").classList.contains("timing-disabled")).toBe(false);
  });

  it("a card with nothing to do is still dimmed and wears nothing", () => {
    const p = live(withDigest());
    expect(p.slot("dead").classList.contains("timing-disabled")).toBe(true);
    expect(p.tile("dead").querySelector(".ready-pip")).toBeNull();
    expect(p.tile("dead").classList.contains("ready")).toBe(false);
  });

  it("tapping the star opens a popover whose foretell row sends the admin menu's payload", () => {
    const p = live(withDigest());
    p.pip("dwarf", "star")!.click();
    flushSync();
    const rows = p.menuRows("dwarf");
    expect(rows).toHaveLength(1);
    expect(rows[0].dataset.special).toBe("special-foretell");
    expect(rows[0].classList.contains("ready")).toBe(true);
    expect(rows[0].disabled).toBe(false);
    rows[0].click();
    flushSync();
    expect(p.sendAction).toHaveBeenCalledWith(
      "special_action",
      { card_id: "dwarf", kind: "foretell", strict: true, auto_tap: true },
      ME,
    );
    // Choosing a row closes the popover.
    expect(p.menuRows("dwarf")).toHaveLength(0);
  });
});

describe("a face-down permanent", () => {
  it("can be turned face up: the star and the ring, and a row for it", () => {
    const p = live(withDigest());
    expect(p.pip("morph", "star")).not.toBeNull();
    expect(p.tile("morph").classList.contains("ready")).toBe(true);
    p.pip("morph", "star")!.click();
    flushSync();
    const rows = p.menuRows("morph");
    expect(rows.map((b) => b.dataset.special)).toEqual(["special-turn_face_up"]);
    rows[0].click();
    expect(p.sendAction).toHaveBeenCalledWith(
      "special_action",
      { card_id: "morph", kind: "turn_face_up", strict: true, auto_tap: true },
      ME,
    );
  });
});

describe("the setting and the digest", () => {
  it("highlights off: no pip, no ring, no accent; the rows and the dim rule stay", () => {
    const view = withDigest();
    const all = legalActionsOf(view);
    const p = mountPanel(view, visibleHighlights(all, false), all);
    expect(p.container.querySelector(".ready-pip")).toBeNull();
    expect(p.container.querySelector(".card.ready")).toBeNull();
    // Dimming is the negative half; the setting never moves it.
    expect(p.slot("dwarf").classList.contains("timing-disabled")).toBe(false);
    expect(p.slot("dead").classList.contains("timing-disabled")).toBe(true);
    // The touch route is the right-click alone now; the row is there.
    const rows = p.rightClick("dwarf");
    expect(rows.map((b) => b.dataset.special)).toEqual(["special-foretell"]);
    expect(rows[0].classList.contains("ready")).toBe(false);
    expect(rows[0].disabled).toBe(false);
  });

  it("the exact digest greys an available row it leaves out, even with highlights off", () => {
    const view = gameView({
      legal_actions: { pass: true, sources: {} },
      legal_moves: [{ type: "pass_priority", player: ME, kind: "pass", label: "pass" }],
    });
    const all = legalActionsOf(view);
    const p = mountPanel(view, visibleHighlights(all, false), all);
    const rows = p.rightClick("dwarf");
    expect(rows[0].disabled).toBe(true);
    expect(rows[0].title).toBe("Can't do this right now");
  });

  it("no digest: no pip, and no row greyed that the server's `available` does not grey", () => {
    const p = live(gameView());
    expect(p.container.querySelector(".ready-pip")).toBeNull();
    expect(p.container.querySelector(".card.ready")).toBeNull();
    const rows = p.rightClick("dwarf");
    expect(rows).toHaveLength(1);
    expect(rows[0].disabled).toBe(false);
    expect(rows[0].classList.contains("ready")).toBe(false);
    const morphRows = p.rightClick("morph");
    expect(morphRows.every((b) => !b.disabled)).toBe(true);
  });
});

describe("a pip is the touch route into the popover", () => {
  // A castable permanent with a live ability: the card's own click
  // means something (tap, cast, select), so the pip must not fire it.
  const engine = mine("engine", "Twin Engine", "Artifact", {
    activated_abilities: [ability(0, "{1}: Scry 1")],
  });
  const legal = legalActionsOf(
    gameView({
      legal_actions: {
        pass: true,
        sources: { engine: { kinds: ["activate"], moves: 1, abilities: ["own:0"] } },
      },
    }),
  );
  const mountCard = () => {
    const onClick = vi.fn();
    const r = render(Card, {
      card: engine,
      pips: readyPips(legal, engine, "battlefield"),
      ready: true,
      legal,
      legalGate: legal,
      onActivateAbility: () => {},
      onClick,
    });
    const pip = r.container.querySelector<HTMLButtonElement>('.ready-pip[data-pip="bolt"]')!;
    const menu = () => r.container.querySelector(".mana-menu");
    return { ...r, onClick, pip, menu };
  };

  it("is a labelled button", () => {
    const { pip } = mountCard();
    expect(pip.tagName).toBe("BUTTON");
    expect(pip.getAttribute("aria-label")).toBe("Activate an ability — open actions");
    expect(pip.getAttribute("aria-haspopup")).toBe("menu");
  });

  it("a click opens the popover and does not fire the card's click", () => {
    const c = mountCard();
    c.pip.click();
    flushSync();
    expect(c.menu()).not.toBeNull();
    expect(c.onClick).not.toHaveBeenCalled();
    expect(c.pip.getAttribute("aria-expanded")).toBe("true");
  });

  it("Enter on a focused pip opens the popover, and the card's Enter does not fire", () => {
    const c = mountCard();
    c.pip.focus();
    expect(document.activeElement).toBe(c.pip);
    c.pip.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
    flushSync();
    expect(c.menu()).not.toBeNull();
    expect(c.onClick).not.toHaveBeenCalled();
  });

  it("Space opens it too", () => {
    const c = mountCard();
    c.pip.dispatchEvent(new KeyboardEvent("keydown", { key: " ", bubbles: true }));
    flushSync();
    expect(c.menu()).not.toBeNull();
  });

  it("a second tap keeps it open; a tap on the card closes it without the card's click", () => {
    const c = mountCard();
    c.pip.click();
    c.pip.click();
    flushSync();
    expect(c.menu()).not.toBeNull();
    c.container.querySelector<HTMLElement>(".card")!.click();
    flushSync();
    expect(c.menu()).toBeNull();
    expect(c.onClick).not.toHaveBeenCalled();
  });
});
