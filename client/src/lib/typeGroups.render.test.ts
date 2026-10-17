// @vitest-environment jsdom
//
// typeGroups.render.test.ts — #4087. Every permanent type has a group of
// its own on a seat's board (creatures, lands, planeswalkers, battles,
// artifacts, enchantments), each with its label, and an empty group
// takes no space. Planeswalkers and battles never overlap or fold their
// cards. The layout itself (widths, shares, the loyalty pip's size, 2/3/4
// seats, 390px) is measured in a real browser by
// tests-e2e/tests/type-groups-4087.spec.ts; jsdom has no layout, so this
// test pins the markup:
//
//   - each type sits in its own labelled list, and a permanent with
//     several types is drawn once, in the first group that applies;
//   - only the planeswalker and battle rows are marked `spread`, which
//     is what turns the overlap off;
//   - a group with no card is absent, so a board without one draws the
//     rows it always did;
//   - two walkers with the same name stay two cards, never one pile.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import type { CardView, GameView, PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

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

const sorin = (id: string, loyalty: number) =>
  permanent(id, "Sorin, Vengeful Bloodlord", "Legendary Planeswalker — Sorin", {
    counters: { loyalty },
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

const gameView = (cards: CardView[]): GameView =>
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
  }) as unknown as GameView;

function mountPanel(cards: CardView[], onCardClick: (id: string) => void = () => {}) {
  const view = gameView(cards);
  return render(
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
      controlledCards: cards,
      exile: zone("exile", undefined),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: onCardClick,
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: () => {},
      onManaAbilityCost: () => {},
    } as never,
  );
}

const names = (container: HTMLElement, label: string): string[] =>
  [...container.querySelectorAll(`[role="list"][aria-label="${label}"] .card`)].map(
    (el) => el.getAttribute("data-instance-id") ?? "",
  );

describe("the per-type groups (#4087)", () => {
  const rock = permanent("rock", "Sol Ring", "Artifact");
  const anthem = permanent("anthem", "Glorious Anthem", "Enchantment");
  const siege = permanent("siege", "Invasion of Ikoria", "Battle — Siege", {
    counters: { defense: 5 },
  });

  it("draws each type in its own labelled list", () => {
    const { container } = mountPanel([
      rock,
      sorin("s1", 4),
      anthem,
      siege,
      sorin("s2", 6),
      permanent("elf", "Llanowar Elves", "Creature — Elf", { power: 1, toughness: 1 }),
      permanent("forest", "Forest", "Basic Land — Forest"),
    ]);
    expect(names(container, "planeswalkers")).toEqual(["s1", "s2"]);
    expect(names(container, "battles")).toEqual(["siege"]);
    expect(names(container, "artifacts")).toEqual(["rock"]);
    expect(names(container, "enchantments")).toEqual(["anthem"]);
    expect(names(container, "creatures")).toEqual(["elf"]);
    expect(names(container, "lands")).toEqual(["forest"]);
  });

  it("draws a permanent with several types once, in the first group that applies", () => {
    const golem = permanent("golem", "Solemn Simulacrum", "Artifact Creature — Golem", {
      power: 2,
      toughness: 2,
    });
    const satyr = permanent("satyr", "Eidolon of Blood", "Enchantment Creature — Satyr", {
      power: 3,
      toughness: 1,
    });
    const arbor = permanent("arbor", "Dryad Arbor", "Land Creature — Forest Dryad", {
      power: 1,
      toughness: 1,
    });
    const synod = permanent("synod", "Seat of the Synod", "Artifact Land");
    const bident = permanent("bident", "Bident of Thassa", "Legendary Enchantment Artifact");
    const { container } = mountPanel([golem, satyr, arbor, synod, bident]);
    expect(names(container, "creatures")).toEqual(["golem", "satyr", "arbor"]);
    expect(names(container, "lands")).toEqual(["synod"]);
    expect(names(container, "artifacts")).toEqual(["bident"]);
    // Nothing is drawn twice.
    expect(container.querySelectorAll(".card[data-instance-id]").length).toBe(5);
  });

  it("marks only the planeswalker and battle rows spread", () => {
    const { container } = mountPanel([sorin("s1", 4), siege, rock, anthem]);
    const spread = [...container.querySelectorAll(".row.spread")].map((r) =>
      r.getAttribute("data-zone"),
    );
    expect(spread).toEqual(["planeswalkers", "battles"]);
  });

  it("draws no row for a type the seat controls none of", () => {
    const { container } = mountPanel([rock]);
    for (const label of ["planeswalkers", "battles", "enchantments", "other"]) {
      expect(container.querySelector(`[role="list"][aria-label="${label}"]`)).toBeNull();
    }
    expect(container.querySelector(".row.spread")).toBeNull();
    expect(names(container, "artifacts")).toEqual(["rock"]);
  });

  it("keeps two walkers of one name as two cards, never one pile", () => {
    const { container } = mountPanel([sorin("s1", 4), sorin("s2", 6)]);
    expect(names(container, "planeswalkers")).toEqual(["s1", "s2"]);
    expect(container.querySelector('[data-testid="group-count"]')).toBeNull();
    const loyalty = [...container.querySelectorAll(".badge.loyalty")].map((b) =>
      b.textContent?.trim(),
    );
    expect(loyalty).toEqual(["4", "6"]);
  });
});
