// @vitest-environment jsdom
//
// legalHighlightsA11y.render.test.ts — ADR 0105 §7, sub-PR 6 (#1789).
// The spoken half of the ready highlights, which only markup can show:
//
//   - a ready card's accessible name gains what it is ready for
//     ("castable", "playable land", "has an ability you can activate",
//     "can be foretold", "can attack", "can block"), and a card that is
//     not ready keeps its bare name;
//   - each pip button is named for what it opens onto, and the combat
//     pips (sword, shield) are drawing only: hidden from the
//     accessibility tree, and not buttons;
//   - a ready popover row's name gains "available";
//   - the action dock header's live region says "N actions available" once
//     when the decision arrives, not again on a later frame of the same
//     decision, and again after the decision has ended and come back;
//   - highlights off say nothing, as they draw nothing.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import Card from "./components/board/Card.svelte";
import ActionDock from "./components/board/ActionDock.svelte";
import {
  actionableCount,
  legalActionsOf,
  readyPips,
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
const OPP = "opp";

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
const mana = (index: number, extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index,
  ref: `own:${index}`,
  ...extra,
});

const hand = () => [
  mine("bolt", "Lightning Bolt", "Instant", { mana_cost: "{R}" }),
  mine("wurm", "Craw Wurm", "Creature — Wurm", { mana_cost: "{4}{G}{G}" }),
  mine("forest", "Forest", "Basic Land — Forest"),
];
const board = () => [
  // Two rows, one live: the bolt pip has no count, and the menu has
  // one ready row and one that is not.
  mine("engine", "Twin Engine", "Artifact", {
    activated_abilities: [ability(0, "{1}: Scry 1"), ability(1, "{9}: Draw")],
  }),
  mine("vivi", "Vivi Ornitier", "Legendary Creature — Wizard", {
    mana_abilities: [mana(0, { label: "Add X mana" })],
  }),
  mine("mountain", "Mountain", "Basic Land — Mountain", {
    mana_abilities: [mana(0, { tap_cost: true, produced: "{R}", label: "Add {R}" })],
  }),
];

const mainDigest: LegalActionsView = {
  pass: true,
  sources: {
    bolt: { kinds: ["cast"], moves: 1, zones: ["hand"] },
    forest: { kinds: ["land"], moves: 1, zones: ["hand"] },
    engine: { kinds: ["activate"], moves: 1, abilities: ["own:0"] },
    vivi: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
    mountain: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
  },
};

const seat = (id: string, n: number, handCards: CardView[] = []): PlayerView =>
  ({
    id,
    name: id,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, handCards),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (
  battlefield: CardView[],
  extra: Partial<GameView> = {},
  step = "precombat_main",
): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, 0, hand()), seat(OPP, 1)],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step },
    mulligans_open: false,
    ...extra,
  }) as unknown as GameView;

const lookups = (view: GameView, highlights = true) => {
  const all = legalActionsOf(view);
  return { legal: visibleHighlights(all, highlights), legalGate: all };
};

function mountPanel(
  view: GameView,
  l: { legal: LegalActions; legalGate: LegalActions },
  combatMode: "idle" | "attack" | "block" = "idle",
) {
  const r = render(
    PlayerPanel as never,
    {
      seat: view.seats[0],
      isSelf: true,
      isActive: true,
      hasPriority: true,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: view.battlefield.cards.filter((c) => c.controller === ME),
      exile: zone("exile", undefined),
      combatMode,
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: () => {},
      onManaAbilityCost: () => {},
      legal: l.legal,
      legalGate: l.legalGate,
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  const name = (id: string) => tile(id).getAttribute("aria-label");
  const pip = (id: string, kind: string) =>
    tile(id).querySelector<HTMLElement>(`.ready-pip[data-pip="${kind}"]`);
  const announced = () =>
    r.container.querySelector<HTMLElement>("[data-ready-announcer]")?.textContent ?? null;
  const setFrame = (v: GameView, highlights = true) =>
    r.setProps({ view: v, ...lookups(v, highlights) } as never);
  return { ...r, tile, name, pip, announced, setFrame };
}

describe("a ready card's accessible name says what it is ready for", () => {
  const view = gameView(board(), { legal_actions: mainDigest });

  it("hand: castable, playable land, and a dead card's bare name", () => {
    const p = mountPanel(view, lookups(view));
    expect(p.name("bolt")).toBe("Lightning Bolt, castable");
    expect(p.name("forest")).toBe("Forest, playable land");
    expect(p.name("wurm")).toBe("Craw Wurm");
  });

  it("battlefield: a live ability; a drop pip alone adds no ring and no phrase; a land says nothing", () => {
    const p = mountPanel(view, lookups(view));
    expect(p.name("engine")).toBe("Twin Engine, has an ability you can activate");
    expect(p.name("vivi")).toBe("Vivi Ornitier");
    expect(p.name("mountain")).toBe("Mountain");
  });

  it("highlights off: every card keeps its bare name", () => {
    const p = mountPanel(view, lookups(view, false));
    expect(p.name("bolt")).toBe("Lightning Bolt");
    expect(p.name("forest")).toBe("Forest");
    expect(p.name("engine")).toBe("Twin Engine");
  });
});

describe("pips are named for what they open onto", () => {
  const view = gameView(board(), { legal_actions: mainDigest });

  it("the bolt and the drop", () => {
    const p = mountPanel(view, lookups(view));
    expect(p.pip("engine", "bolt")?.getAttribute("aria-label")).toBe(
      "Activate an ability — open actions",
    );
    expect(p.pip("vivi", "drop")?.getAttribute("aria-label")).toBe(
      "Mana ability available — open actions",
    );
  });

  it("the star names the special action, and the card says it", () => {
    const dwarf = mine("dwarf", "Dwarven Hammer", "Artifact — Equipment", {
      special_actions: [{ kind: "foretell", label: "Foretell {2}", cost: "{2}", available: true }],
    } as Partial<CardView>);
    const legal = legalActionsOf(
      gameView([], {
        legal_actions: {
          pass: true,
          sources: {
            dwarf: { kinds: ["special_action"], moves: 1, special_actions: ["foretell"] },
          },
        },
      }),
    );
    const r = render(Card, {
      card: dwarf,
      pips: readyPips(legal, dwarf, "hand"),
      ready: true,
      readyZone: "hand" as const,
      legal,
      legalGate: legal,
      onSpecialAction: () => {},
    });
    const star = r.container.querySelector<HTMLElement>('.ready-pip[data-pip="star"]')!;
    expect(star.getAttribute("aria-label")).toBe("Foretell available — open actions");
    expect(r.container.querySelector(".card")?.getAttribute("aria-label")).toBe(
      "Dwarven Hammer, can be foretold",
    );
  });
});

describe("a ready popover row's name gains 'available'", () => {
  it("the live row says it; the one the server would refuse does not", () => {
    const view = gameView(board(), { legal_actions: mainDigest });
    const p = mountPanel(view, lookups(view));
    p.tile("engine").dispatchEvent(
      new MouseEvent("contextmenu", { bubbles: true, cancelable: true }),
    );
    flushSync();
    const rows = [...p.tile("engine").querySelectorAll<HTMLButtonElement>(".mana-menu .menu-item")];
    const scry = rows.find((b) => b.textContent?.includes("Scry 1"))!;
    const draw = rows.find((b) => b.textContent?.includes("Draw"))!;
    expect(scry.classList.contains("ready")).toBe(true);
    expect(scry.querySelector(".sr-only")?.textContent).toBe(", available");
    expect(draw.querySelector(".sr-only")).toBeNull();
  });
});

describe("combat: the sword and the shield", () => {
  const bear = mine("bear", "Grizzly Bears", "Creature — Bear", { power: 2, toughness: 2 });
  const tired = mine("tired", "Tired Bear", "Creature — Bear", { tapped: true });

  it("an attack candidate wears a sword that is drawing only, and its name says it can attack", () => {
    const view = gameView(
      [bear, tired],
      {
        legal_actions: {
          pass: false,
          sources: { bear: { kinds: ["attack"], moves: 1, attack_targets: [OPP] } },
        },
      },
      "declare_attackers",
    );
    const p = mountPanel(view, lookups(view), "attack");
    const sword = p.pip("bear", "sword")!;
    expect(sword).not.toBeNull();
    expect(sword.tagName).toBe("SPAN");
    expect(sword.getAttribute("aria-hidden")).toBe("true");
    expect(p.name("bear")).toBe("Grizzly Bears, can attack");
    expect(p.pip("tired", "sword")).toBeNull();
    expect(p.name("tired")).toBe("Tired Bear");
  });

  it("a block candidate wears a shield", () => {
    const giant = {
      ...mine("giant", "Hill Giant", "Creature — Giant"),
      owner: OPP,
      controller: OPP,
      attacking_target: ME,
    } as CardView;
    const view = gameView(
      [bear, giant],
      {
        legal_actions: {
          sources: { bear: { kinds: ["block"], moves: 1, blocks: ["giant"] } },
        },
      },
      "declare_blockers",
    );
    const p = mountPanel(view, lookups(view), "block");
    expect(p.pip("bear", "shield")?.getAttribute("aria-hidden")).toBe("true");
    expect(p.pip("bear", "sword")).toBeNull();
    expect(p.name("bear")).toBe("Grizzly Bears, can block");
  });

  it("highlights off: no combat pip", () => {
    const view = gameView(
      [bear],
      {
        legal_actions: {
          pass: false,
          sources: { bear: { kinds: ["attack"], moves: 1, attack_targets: [OPP] } },
        },
      },
      "declare_attackers",
    );
    const p = mountPanel(view, lookups(view, false), "attack");
    expect(p.pip("bear", "sword")).toBeNull();
  });
});

// Since ADR 0111 PR 2 the live region is in the action dock's header,
// and Game.svelte counts the viewer's ready cards the way the self panel
// used to: actionableCount over the highlight lookup.
function mountDock(view: GameView, l: { legal: LegalActions }) {
  const count = (v: GameView, legal: LegalActions) => actionableCount(legal, v, ME);
  const r = render(
    ActionDock as never,
    {
      view,
      viewerHasPriority: true,
      viewerIsActive: true,
      autopassEnabled: false,
      readyActions: count(view, l.legal),
      onPassPriority: () => {},
      onPassTurn: () => {},
      onToggleAutopass: () => {},
    } as never,
  );
  const announced = () =>
    r.container.querySelector<HTMLElement>("[data-ready-announcer]")?.textContent ?? null;
  const setFrame = (v: GameView, highlights = true) =>
    r.setProps({ view: v, readyActions: count(v, lookups(v, highlights).legal) } as never);
  return { ...r, announced, setFrame };
}

describe("the live region: 'N actions available', once per arrival", () => {
  // Lightning Bolt, Forest, Twin Engine and Vivi (a drop pip); not the
  // Craw Wurm, and not the Mountain's {T} mana.
  const arrive = () => gameView(board(), { legal_actions: mainDigest });
  // The same decision, a frame later: the land was played, so the
  // digest changed, and priority never left the seat.
  const later = () =>
    gameView(board(), {
      legal_actions: {
        pass: true,
        sources: {
          bolt: { kinds: ["cast"], moves: 1, zones: ["hand"] },
          engine: { kinds: ["activate"], moves: 1, abilities: ["own:0"] },
        },
      },
    });
  // Priority moved on: no digest.
  const gone = () =>
    gameView(board(), { turn: { number: 3, active_seat: 0, priority_holder: 1 } as never });

  it("announces when the decision arrives", () => {
    const p = mountDock(arrive(), lookups(arrive()));
    expect(p.announced()).toBe("4 actions available");
  });

  it("says nothing new on a later frame of the same decision", () => {
    const p = mountDock(arrive(), lookups(arrive()));
    p.setFrame(later());
    expect(p.announced()).toBe("4 actions available");
  });

  it("empties when priority leaves, and announces again when it comes back", () => {
    const p = mountDock(gone(), lookups(gone()));
    expect(p.announced()).toBe("");
    p.setFrame(arrive());
    expect(p.announced()).toBe("4 actions available");
    p.setFrame(gone());
    expect(p.announced()).toBe("");
    p.setFrame(later());
    expect(p.announced()).toBe("2 actions available");
  });

  it("highlights off, or autopass passing the frame: says nothing", () => {
    const p = mountDock(arrive(), lookups(arrive(), false));
    expect(p.announced()).toBe("");
  });

  it("is a polite status region", () => {
    const p = mountDock(arrive(), lookups(arrive()));
    const region = p.container.querySelector("[data-ready-announcer]")!;
    expect(region.getAttribute("role")).toBe("status");
    expect(region.getAttribute("aria-live")).toBe("polite");
  });
});
