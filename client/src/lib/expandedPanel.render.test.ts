// @vitest-environment jsdom
//
// expandedPanel.render.test.ts — ADR 0120 PR 2, the plumbing for the
// expanded board. Nothing mounts an expanded panel yet; these tests
// mount one beside the table's to pin what PR 3's overlay relies on:
//
//   - `expanded` makes the panel an unnamed `group`, upright, with no
//     dock or coach cell, so "your board" / "<name> board" stays unique;
//   - the ability popover records the surface it was opened from, and
//     only the Card on that surface draws it, so two copies of a card
//     draw it once (ADR 0120 §3, "One popover");
//   - a token group swaps in the popover's member only on that surface.

import { afterEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import { abilityPopover, openAbilityPopover } from "./abilityPopover";
import type { CardView, GameView, PlayerView } from "./protocol";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
});

const ME = "me";
const BOB = "bob";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, n: number): PlayerView =>
  ({
    id,
    name: id === ME ? "Me" : "Bob",
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

const gameView = (battlefield: CardView[]): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, 0), seat(BOB, 1)],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "precombat_main" },
    mulligans_open: false,
  }) as unknown as GameView;

// Two usable rows, one not mana: a left-click opens the popover.
const relic = (): CardView =>
  ({
    instance_id: "relic",
    name: "Relic of Sauron",
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Legendary Artifact",
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
  }) as unknown as CardView;

function mountPanel(view: GameView, extra: Record<string, unknown> = {}) {
  const isSelf = (extra.isSelf as boolean | undefined) ?? true;
  return render(
    PlayerPanel as never,
    {
      seat: isSelf ? view.seats[0] : view.seats[1],
      isSelf,
      isActive: isSelf,
      hasPriority: isSelf,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: isSelf ? view.battlefield.cards : [],
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
      ...extra,
    } as never,
  );
}

const panelOf = (c: HTMLElement) => c.querySelector<HTMLElement>(".panel")!;
const tileIn = (c: HTMLElement) => c.querySelector<HTMLElement>('.card[data-instance-id="relic"]')!;
const menusIn = (c: HTMLElement) => c.querySelectorAll(".mana-menu");

describe("PlayerPanel's `expanded` prop", () => {
  it("is a region named for the seat without it, as before", () => {
    const view = gameView([]);
    expect(panelOf(mountPanel(view).container).getAttribute("role")).toBe("region");
    expect(panelOf(mountPanel(view).container).getAttribute("aria-label")).toBe("your board");
    const bob = panelOf(mountPanel(view, { isSelf: false }).container);
    expect(bob.getAttribute("aria-label")).toBe("Bob board");
  });

  it("makes the panel an unnamed group, so the board's name stays unique", () => {
    const view = gameView([]);
    mountPanel(view);
    mountPanel(view, { expanded: true });
    mountPanel(view, { isSelf: false });
    mountPanel(view, { isSelf: false, expanded: true });
    const named = (name: string) =>
      [...document.querySelectorAll("[aria-label]")].filter((e) =>
        e.getAttribute("aria-label")!.includes(name),
      );
    expect(named("your board")).toHaveLength(1);
    expect(named("Bob board")).toHaveLength(1);
    const groups = [...document.querySelectorAll<HTMLElement>(".panel.expanded")];
    expect(groups).toHaveLength(2);
    for (const g of groups) {
      expect(g.getAttribute("role")).toBe("group");
      expect(g.hasAttribute("aria-label")).toBe(false);
    }
  });

  it("is upright and leaves the dock's and the coach's cells to the table panel", () => {
    const view = gameView([]);
    const p = mountPanel(view, { expanded: true, flipped: true, docked: true, coached: true });
    const panel = panelOf(p.container);
    expect(panel.classList.contains("flipped")).toBe(false);
    expect(panel.classList.contains("docked")).toBe(false);
    expect(p.container.querySelector(".dock-spacer")).toBeNull();
    expect(p.container.querySelector(".coach-spacer")).toBeNull();
    // The same props on a table panel keep all three.
    const table = mountPanel(view, { flipped: true, docked: true, coached: true });
    expect(panelOf(table.container).classList.contains("flipped")).toBe(true);
    expect(table.container.querySelector(".dock-spacer")).not.toBeNull();
    expect(table.container.querySelector(".coach-spacer")).not.toBeNull();
  });
});

describe("the ability popover's surface (ADR 0120 §3)", () => {
  it("defaults to the table", () => {
    openAbilityPopover("relic");
    expect(get(abilityPopover)).toEqual({ cardID: "relic", surface: "table" });
  });

  it("a left-click in the expanded copy opens it there, and only there", () => {
    const view = gameView([relic()]);
    const table = mountPanel(view);
    const overlay = mountPanel(view, { expanded: true });
    click(tileIn(overlay.container));
    expect(get(abilityPopover)).toEqual({ cardID: "relic", surface: "expanded" });
    expect(menusIn(overlay.container)).toHaveLength(1);
    expect(menusIn(table.container)).toHaveLength(0);
  });

  it("a left-click on the table copy opens it on the table, and only there", () => {
    const view = gameView([relic()]);
    const table = mountPanel(view);
    const overlay = mountPanel(view, { expanded: true });
    click(tileIn(table.container));
    expect(get(abilityPopover)).toEqual({ cardID: "relic", surface: "table" });
    expect(menusIn(table.container)).toHaveLength(1);
    expect(menusIn(overlay.container)).toHaveLength(0);
  });

  it("a right-click records its copy's surface too", () => {
    const view = gameView([relic()]);
    const table = mountPanel(view);
    const overlay = mountPanel(view, { expanded: true });
    tileIn(overlay.container).dispatchEvent(
      new MouseEvent("contextmenu", { bubbles: true, cancelable: true }),
    );
    flushSync();
    expect(get(abilityPopover)).toEqual({ cardID: "relic", surface: "expanded" });
    expect(menusIn(overlay.container)).toHaveLength(1);
    expect(menusIn(table.container)).toHaveLength(0);
  });

  it("a click on the other copy moves it rather than closing it", () => {
    const view = gameView([relic()]);
    const table = mountPanel(view);
    const overlay = mountPanel(view, { expanded: true });
    click(tileIn(overlay.container));
    click(tileIn(table.container));
    expect(get(abilityPopover)).toEqual({ cardID: "relic", surface: "table" });
    expect(menusIn(table.container)).toHaveLength(1);
    expect(menusIn(overlay.container)).toHaveLength(0);
    // And a click on the copy that draws it closes it, as before.
    click(tileIn(table.container));
    expect(get(abilityPopover)).toBeNull();
  });
});

describe("a token group follows the popover only on its surface", () => {
  const clue = (id: string): CardView =>
    ({
      instance_id: id,
      name: "Clue",
      owner: ME,
      controller: ME,
      type_line: "Token Artifact — Clue",
      battle_x: Number(id.slice(1)),
      activated_abilities: [
        { index: 0, ref: "own:0", label: "{2}, Sacrifice this artifact: Draw a card." },
        { index: 1, ref: "own:1", label: "{1}: Scry 1.", mana_cost: "{1}" },
      ],
    }) as unknown as CardView;

  it("draws the chosen member on the popover's surface and its first member elsewhere", () => {
    const cards = [clue("c1"), clue("c2"), clue("c3")];
    const r = render(
      BattlefieldRow as never,
      {
        label: "creatures",
        cards,
        viewerID: ME,
        attachmentsByHost: {},
        onCardClick: () => {},
        onGroupClick: () => {},
      } as never,
    );
    const drawn = () =>
      r.container.querySelector<HTMLElement>(".pile.group .card")?.dataset.instanceId;
    expect(drawn()).toBe("c1");
    openAbilityPopover("c2", "expanded");
    flushSync();
    expect(drawn()).toBe("c1");
    openAbilityPopover("c2", "table");
    flushSync();
    expect(drawn()).toBe("c2");
  });
});
