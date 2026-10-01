// @vitest-environment jsdom
//
// legalHighlightsCombat.render.test.ts — ADR 0105 sub-PR 5 (#1789).
// What the board draws during a combat declaration from
// legalActions.ts, which only markup can show:
//
//   - declare attackers: each creature the server would let attack
//     wears the ready ring, and one it would not does not;
//   - with an attacker selected, its defenders light: the players in
//     its `attack_targets` keep the identity's green ring (and only
//     those), the planeswalkers and battles get the ready ring, and a
//     click on one declares the attack;
//   - declare blockers: each creature that may block wears the ring;
//     with one selected, the attackers it may block wear it too, drawn
//     over their red ring;
//   - an already-declared attacker or blocker keeps its red or blue
//     ring and no ready ring;
//   - highlights off take every combat ring away and leave the gates;
//     no digest lights nothing and withholds nothing new.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import {
  NO_LEGAL_ACTIONS,
  legalActionsOf,
  visibleHighlights,
  type LegalActions,
} from "./legalActions";
import type { CardView, GameView, LegalActionsView, PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  localStorage.clear();
});

const ME = "me";
const OPP = "opp";
const OPP2 = "opp2";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const permanent = (
  id: string,
  controller: string,
  typeLine: string,
  extra: Partial<CardView> = {},
): CardView =>
  ({
    instance_id: id,
    name: id,
    owner: controller,
    controller,
    known_by_you: true,
    type_line: typeLine,
    ...extra,
  }) as CardView;

const creature = (id: string, controller: string, extra: Partial<CardView> = {}) =>
  permanent(id, controller, "Creature — Bear", { power: 2, toughness: 2, ...extra });

const seat = (id: string, n: number): PlayerView =>
  ({
    id,
    name: id,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (
  cards: CardView[],
  step: string,
  activeSeat: number,
  extra: Partial<GameView> = {},
): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, 0), seat(OPP, 1), seat(OPP2, 2)],
    battlefield: zone("battlefield", undefined, cards),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      number: 3,
      active_seat: activeSeat,
      priority_holder: activeSeat,
      phase: "combat",
      step,
    },
    mulligans_open: false,
    ...extra,
  }) as unknown as GameView;

interface MountOpts {
  seatID: string;
  mode: "idle" | "attack" | "block";
  selected?: string | null;
  legal: LegalActions;
  legalGate: LegalActions;
  onDeclareAttack?: (id: string) => void;
}

// mountPanel mounts one seat's panel the way Board does: every
// non-spectator panel gets the viewer's two lookups.
function mountPanel(view: GameView, o: MountOpts) {
  const s = view.seats.find((x) => x.id === o.seatID)!;
  const r = render(
    PlayerPanel as never,
    {
      seat: s,
      isSelf: o.seatID === ME,
      isActive: false,
      hasPriority: false,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: view.battlefield.cards.filter((c) => c.controller === o.seatID),
      exile: zone("exile", undefined),
      combatMode: o.mode,
      selectedCombatCardID: o.selected ?? null,
      onSelectCombatCard: () => {},
      onDeclareAttack: o.onDeclareAttack ?? (() => {}),
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: () => {},
      onManaAbilityCost: () => {},
      legal: o.legal,
      legalGate: o.legalGate,
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  const ready = (id: string) => tile(id).classList.contains("ready");
  const identityTargetable = () =>
    r.container.querySelector(".identity")!.classList.contains("targetable");
  return { ...r, tile, ready, identityTargetable };
}

// Lookups for a frame: highlights on, highlights off, or none at all.
const on = (v: GameView) => {
  const all = legalActionsOf(v);
  return { legal: visibleHighlights(all, true), legalGate: all };
};
const off = (v: GameView) => {
  const all = legalActionsOf(v);
  return { legal: visibleHighlights(all, false), legalGate: all };
};

// ---- declare attackers, the viewer's turn ----------------------------

const attackBoard = () => [
  creature("bear", ME),
  // Untapped and unflagged, but the server leaves it out (a tax it
  // cannot pay): the old re-derivation would have called it ready.
  creature("taxed", ME),
  creature("sick", ME, { summoning_sick: true }),
  creature("declared", ME, { attacking_target: OPP, tapped: true }),
  permanent("walker", OPP, "Legendary Planeswalker — Jace", { counters: { loyalty: 3 } }),
  permanent("other-walker", OPP2, "Legendary Planeswalker — Chandra", {
    counters: { loyalty: 4 },
  }),
];
// The bear may attack the opponent or their planeswalker; nobody else.
const attackDigest: LegalActionsView = {
  pass: true,
  sources: {
    bear: { kinds: ["attack"], moves: 2, attack_targets: [OPP, "walker"] },
  },
};
const attackView = () =>
  gameView(attackBoard(), "declare_attackers", 0, { legal_actions: attackDigest });

describe("declare attackers", () => {
  it("rings the creatures the server would let attack, and not the rest", () => {
    const v = attackView();
    const p = mountPanel(v, { seatID: ME, mode: "attack", ...on(v) });
    expect(p.ready("bear")).toBe(true);
    expect(p.ready("taxed")).toBe(false);
    expect(p.ready("sick")).toBe(false);
  });

  it("an already-declared attacker keeps its red ring and no ready ring", () => {
    const v = attackView();
    const p = mountPanel(v, { seatID: ME, mode: "attack", ...on(v) });
    expect(p.tile("declared").classList.contains("attacking")).toBe(true);
    expect(p.ready("declared")).toBe(false);
  });

  it("a selected attacker lights its defenders, and only those", () => {
    const v = attackView();
    const opp = mountPanel(v, { seatID: OPP, mode: "attack", selected: "bear", ...on(v) });
    expect(opp.identityTargetable()).toBe(true);
    expect(opp.ready("walker")).toBe(true);

    const opp2 = mountPanel(v, { seatID: OPP2, mode: "attack", selected: "bear", ...on(v) });
    expect(opp2.identityTargetable()).toBe(false);
    expect(opp2.ready("other-walker")).toBe(false);
  });

  it("nothing lights before an attacker is selected", () => {
    const v = attackView();
    const opp = mountPanel(v, { seatID: OPP, mode: "attack", ...on(v) });
    expect(opp.identityTargetable()).toBe(false);
    expect(opp.ready("walker")).toBe(false);
  });

  it("clicking a lit planeswalker declares the attack on it", () => {
    const v = attackView();
    const onDeclareAttack = vi.fn();
    const opp = mountPanel(v, {
      seatID: OPP,
      mode: "attack",
      selected: "bear",
      onDeclareAttack,
      ...on(v),
    });
    opp.tile("walker").click();
    expect(onDeclareAttack).toHaveBeenCalledWith("walker");
  });

  it("re-pointing an attacker already declared keeps every opponent open", () => {
    const v = attackView();
    const opp2 = mountPanel(v, { seatID: OPP2, mode: "attack", selected: "declared", ...on(v) });
    expect(opp2.identityTargetable()).toBe(true);
  });

  it("highlights off: no combat ring, and the defender gate is unchanged", () => {
    const v = attackView();
    const me = mountPanel(v, { seatID: ME, mode: "attack", ...off(v) });
    expect(me.container.querySelector(".card.ready")).toBeNull();
    const opp = mountPanel(v, { seatID: OPP, mode: "attack", selected: "bear", ...off(v) });
    expect(opp.ready("walker")).toBe(false);
    expect(opp.identityTargetable()).toBe(true);
    const opp2 = mountPanel(v, { seatID: OPP2, mode: "attack", selected: "bear", ...off(v) });
    expect(opp2.identityTargetable()).toBe(false);
  });

  it("no digest: nothing lights and nothing is newly withheld", () => {
    const v = gameView(attackBoard(), "declare_attackers", 0);
    const me = mountPanel(v, { seatID: ME, mode: "attack", ...on(v) });
    expect(me.container.querySelector(".card.ready")).toBeNull();
    const onDeclareAttack = vi.fn();
    const opp2 = mountPanel(v, {
      seatID: OPP2,
      mode: "attack",
      selected: "bear",
      onDeclareAttack,
      ...on(v),
    });
    // Every opponent stays a target, as before the digest.
    expect(opp2.identityTargetable()).toBe(true);
    expect(opp2.ready("other-walker")).toBe(false);
    // And a planeswalker click is what it was: not an attack.
    opp2.tile("other-walker").click();
    expect(onDeclareAttack).not.toHaveBeenCalled();
  });

  it("a spectator's lookups know nothing: nothing lights", () => {
    const v = attackView();
    const opp = mountPanel(v, {
      seatID: OPP,
      mode: "attack",
      selected: "bear",
      legal: NO_LEGAL_ACTIONS,
      legalGate: NO_LEGAL_ACTIONS,
    });
    expect(opp.ready("walker")).toBe(false);
  });
});

// ---- declare blockers, the viewer defending ---------------------------

const blockBoard = () => [
  creature("giant", OPP, { attacking_target: ME, tapped: true }),
  creature("ogre", OPP, { attacking_target: ME, tapped: true }),
  creature("wall", ME),
  creature("vet", ME),
  creature("tired", ME, { tapped: true }),
  // Blocking already and with room for another (#1706): the digest
  // still lists it, but its blue ring is what it wears.
  creature("old", ME, { blocking_target: "ogre" }),
];
const blockDigest: LegalActionsView = {
  sources: {
    wall: { kinds: ["block"], moves: 1, blocks: ["giant"] },
    vet: { kinds: ["block"], moves: 2, blocks: ["giant", "ogre"] },
    old: { kinds: ["block"], moves: 1, blocks: ["giant"] },
  },
};
const blockView = () =>
  gameView(blockBoard(), "declare_blockers", 1, { legal_actions: blockDigest });

describe("declare blockers", () => {
  it("rings the creatures that may block, and not a tapped one", () => {
    const v = blockView();
    const p = mountPanel(v, { seatID: ME, mode: "block", ...on(v) });
    expect(p.ready("wall")).toBe(true);
    expect(p.ready("vet")).toBe(true);
    expect(p.ready("tired")).toBe(false);
  });

  it("an already-declared blocker keeps its blue ring and no ready ring", () => {
    const v = blockView();
    const p = mountPanel(v, { seatID: ME, mode: "block", ...on(v) });
    expect(p.tile("old").classList.contains("blocking")).toBe(true);
    expect(p.ready("old")).toBe(false);
  });

  it("a selected blocker lights the attackers it may block, over their red ring", () => {
    const v = blockView();
    const opp = mountPanel(v, { seatID: OPP, mode: "block", selected: "wall", ...on(v) });
    const giant = opp.tile("giant");
    expect(giant.classList.contains("attacking")).toBe(true);
    expect(giant.classList.contains("ready")).toBe(true);
    expect(giant.classList.contains("combat-target")).toBe(true);
    // The wall may not block the ogre: red ring only.
    expect(opp.tile("ogre").classList.contains("attacking")).toBe(true);
    expect(opp.ready("ogre")).toBe(false);
    expect(opp.tile("ogre").classList.contains("combat-target")).toBe(false);
  });

  it("a different blocker lights its own attackers", () => {
    const v = blockView();
    const opp = mountPanel(v, { seatID: OPP, mode: "block", selected: "vet", ...on(v) });
    expect(opp.ready("giant")).toBe(true);
    expect(opp.ready("ogre")).toBe(true);
  });

  it("highlights off: no combat ring on either side", () => {
    const v = blockView();
    const me = mountPanel(v, { seatID: ME, mode: "block", ...off(v) });
    expect(me.container.querySelector(".card.ready")).toBeNull();
    const opp = mountPanel(v, { seatID: OPP, mode: "block", selected: "vet", ...off(v) });
    expect(opp.container.querySelector(".card.ready")).toBeNull();
  });

  it("no digest: nothing lights", () => {
    const v = gameView(blockBoard(), "declare_blockers", 1);
    const me = mountPanel(v, { seatID: ME, mode: "block", ...on(v) });
    expect(me.container.querySelector(".card.ready")).toBeNull();
    const opp = mountPanel(v, { seatID: OPP, mode: "block", selected: "vet", ...on(v) });
    expect(opp.container.querySelector(".card.ready")).toBeNull();
  });
});
