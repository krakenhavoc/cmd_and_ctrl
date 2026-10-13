// @vitest-environment jsdom
//
// #3280: owner report from live play — "I was not able to cycle a land.
// It showed I could do something, but I did not see anything other than
// saying cannot play a land. Raffine's Tower was one of the cards."
//
// A land in hand with cycling (or landcycling, or channel, or a granted
// zone ability) is lit by the server's digest — it wears the bolt and the
// ring and is not dimmed — but once the turn's land drop is spent it is
// not playable, and the Hand gave such a card no click at all: the only
// words anywhere were the tooltip's "No land plays left this turn". The
// popover was reachable only by a right-click or the pip. A left-click on
// a card that cannot be played but can do something else from the hand
// now opens that popover.
//
//   - land with cycling, drop spent: a click opens the cycling row, plays
//     nothing, and the row sends the activation;
//   - land with cycling, drop available: the click still plays the land
//     (the fast path), and the right-click still has the cycling row;
//   - a nonland with cycling that cannot be cast: the same click opens it;
//     one that can be cast keeps the click as the cast;
//   - a land with nothing to do, drop spent: still dimmed, still inert.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import { legalActionsOf, visibleHighlights } from "./legalActions";
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

const cycling = (): ActivatedAbilityView => ({
  index: 0,
  ref: "own:0",
  label: "Cycling {3} ({3}, Discard this card: Draw a card.)",
  mana_cost: "{3}",
  discard_self: true,
});

const tower = () =>
  mine("tower", "Raffine's Tower", "Land — Plains Island Swamp", {
    zone_abilities: [cycling()],
  });
const plainLand = () => mine("forest", "Forest", "Basic Land — Forest");
const bear = () =>
  mine("bear", "Gnarled Bear", "Creature — Bear", {
    mana_cost: "{1}{G}",
    zone_abilities: [cycling()],
  });

const castMove = (source: string): LegalMoveView => ({
  type: "cast_spell",
  player: ME,
  kind: "cast",
  label: "Play",
  source,
  params: { instance_id: source, from_zone: "hand" },
});
const cycleMove = (source: string): LegalMoveView => ({
  type: "activate_ability",
  player: ME,
  kind: "activate",
  label: "Cycling {3}",
  source,
  params: { card_id: source, ref: "own:0" },
});
const passMove: LegalMoveView = {
  type: "pass_priority",
  player: ME,
  kind: "pass",
  label: "pass",
};

const seat = (hand: CardView[], landsPlayed: number): PlayerView =>
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
    land_drops_per_turn: 1,
    lands_played_this_turn: landsPlayed,
  }) as unknown as PlayerView;

function frame(hand: CardView[], landsPlayed: number, moves: LegalMoveView[]): GameView {
  const sources: NonNullable<LegalActionsView["sources"]> = {};
  for (const m of moves) {
    if (!m.source) continue;
    const s = (sources[m.source] ??= { kinds: [], moves: 0 });
    s.kinds.push(m.kind === "cast" ? "cast" : "activate");
    s.moves += 1;
    if (m.type === "activate_ability") s.abilities = ["own:0"];
  }
  const digest: LegalActionsView = { pass: true, sources };
  return {
    id: "g1",
    state: "active",
    seats: [seat(hand, landsPlayed)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    legal_actions: digest,
    legal_moves: [passMove, ...moves],
  } as unknown as GameView;
}

function mount(view: GameView) {
  const all = legalActionsOf(view);
  const sendAction = vi.fn();
  const onPlayCard = vi.fn();
  const onActivateAbility = vi.fn();
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
      controlledCards: [],
      exile: zone("exile", undefined),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard,
      onDrawCard: () => {},
      onActivateAbility,
      onManaAbilityCost: () => {},
      legal: visibleHighlights(all, true),
      legalGate: all,
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  const slot = (id: string) => tile(id).closest<HTMLElement>(".hand-slot")!;
  const rows = (id: string) => [
    ...tile(id).querySelectorAll<HTMLButtonElement>(
      ".mana-menu .menu-item:not([data-raw-tap]):not([data-cast-anyway])",
    ),
  ];
  const leftClick = (id: string) => {
    tile(id).click();
    flushSync();
  };
  const rightClick = (id: string) => {
    tile(id).dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    flushSync();
  };
  return {
    ...r,
    sendAction,
    onPlayCard,
    onActivateAbility,
    tile,
    slot,
    rows,
    leftClick,
    rightClick,
  };
}

describe("a land in hand with cycling, land drop spent", () => {
  const view = () => frame([tower(), plainLand()], 1, [cycleMove("tower")]);

  it("is lit, not dimmed, and its tooltip points at the abilities", () => {
    const p = mount(view());
    expect(p.tile("tower").classList.contains("ready")).toBe(true);
    expect(p.slot("tower").classList.contains("timing-disabled")).toBe(false);
    expect(p.slot("tower").title).toContain("No land plays left this turn");
    expect(p.slot("tower").title).toContain("other abilities");
  });

  it("a left-click opens the cycling row instead of doing nothing", () => {
    const p = mount(view());
    expect(p.rows("tower")).toHaveLength(0);
    p.leftClick("tower");
    const rows = p.rows("tower");
    expect(rows).toHaveLength(1);
    expect(rows[0].textContent).toContain("Cycling {3}");
    // It is not a play.
    expect(p.onPlayCard).not.toHaveBeenCalled();

    rows[0].click();
    flushSync();
    expect(p.onActivateAbility).toHaveBeenCalledTimes(1);
    expect(p.onActivateAbility.mock.calls[0][0].instance_id).toBe("tower");
    expect(p.onActivateAbility.mock.calls[0][1]).toBe(0);
  });

  it("a land with nothing else to do stays dimmed and inert", () => {
    const p = mount(view());
    expect(p.slot("forest").classList.contains("timing-disabled")).toBe(true);
    expect(p.slot("forest").title).toBe("No land plays left this turn");
    p.leftClick("forest");
    expect(p.rows("forest")).toHaveLength(0);
    expect(p.onPlayCard).not.toHaveBeenCalled();
  });
});

describe("a land in hand with cycling, land drop available", () => {
  const view = () => frame([tower()], 0, [castMove("tower"), cycleMove("tower")]);

  it("keeps the play fast path on a left-click", () => {
    const p = mount(view());
    p.leftClick("tower");
    expect(p.onPlayCard).toHaveBeenCalledTimes(1);
    expect(p.rows("tower")).toHaveLength(0);
    expect(p.slot("tower").title).toBe("");
  });

  it("still has the cycling row on a right-click", () => {
    const p = mount(view());
    p.rightClick("tower");
    const rows = p.rows("tower");
    expect(rows).toHaveLength(1);
    expect(rows[0].textContent).toContain("Cycling {3}");
  });
});

describe("a nonland in hand with cycling", () => {
  it("that cannot be cast opens its cycling row on a left-click", () => {
    const p = mount(frame([bear()], 0, [cycleMove("bear")]));
    p.leftClick("bear");
    expect(p.rows("bear")).toHaveLength(1);
    expect(p.onPlayCard).not.toHaveBeenCalled();
  });

  it("that can be cast keeps the click as the cast", () => {
    const p = mount(frame([bear()], 0, [castMove("bear"), cycleMove("bear")]));
    p.leftClick("bear");
    expect(p.onPlayCard).toHaveBeenCalledTimes(1);
    expect(p.rows("bear")).toHaveLength(0);
  });
});
