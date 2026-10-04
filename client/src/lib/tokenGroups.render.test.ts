// @vitest-environment jsdom
//
// #1724 — identical tokens on the board. What only the markup can say:
// a group draws as at most two cards (untapped, tapped) with a count
// badge, or ONE card when every member shares a state; every hidden
// member keeps an anchor with its instance ID on the group's card; and
// the member list picks one, N or all for an attack, picks one
// specific token for a targeting prompt, and shows each member's
// counters and attachments. The pure rules are in tokenGroups.test.ts.

import { afterEach, describe, expect, it } from "vitest";

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import PlayerPanel from "./components/board/PlayerPanel.svelte";
import TokenGroupModal from "./components/board/TokenGroupModal.svelte";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { targeting, type TargetingState } from "./targeting";
import { modalDepth } from "./modalLayers";
import { abilityPopover } from "./abilityPopover";
import { get } from "svelte/store";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

afterEach(() => {
  targeting.set(null);
  cleanup();
});

function soldier(id: string, extra: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name: "Soldier",
    owner: "me",
    controller: "me",
    type_line: "Token Creature — Soldier",
    colors: ["W"],
    power: 1,
    toughness: 1,
    battle_x: Number(id.replace(/\D/g, "")) || 0,
    ...extra,
  };
}

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}
function seat(id: string, name: string, n: number): PlayerView {
  return {
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}
function snap(cards: CardView[], step = "declare_attackers"): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, cards),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step,
    },
    mulligans_open: false,
  };
}

const text = (el: Element | null | undefined): string =>
  el?.textContent?.replace(/\s+/g, " ").trim() ?? "";

function mountRow(cards: CardView[], attachmentsByHost: Record<string, CardView[]> = {}) {
  const opened: string[] = [];
  const clicked: string[] = [];
  const r = render(
    BattlefieldRow as never,
    {
      label: "creatures",
      cards,
      viewerID: "me",
      attachmentsByHost,
      onCardClick: (c: CardView) => clicked.push(c.instance_id),
      onGroupClick: (k: string) => opened.push(k),
    } as never,
  );
  return { container: r.container, opened, clicked };
}

const badges = (c: HTMLElement) =>
  [...c.querySelectorAll('[data-testid="group-count"]')].map((b) => text(b));
const drawnCards = (c: HTMLElement) =>
  [...c.querySelectorAll<HTMLElement>(".card[data-instance-id]")].map((e) => e.dataset.instanceId);

describe("a token group on the board (#1724)", () => {
  it("draws two cards, untapped and tapped, each with its count", () => {
    const cards = [
      soldier("s1"),
      soldier("s2", { tapped: true }),
      soldier("s3"),
      soldier("s4", { tapped: true }),
      soldier("s5"),
    ];
    const { container } = mountRow(cards);
    expect(drawnCards(container)).toEqual(["s1", "s2"]);
    expect(badges(container)).toEqual(["×3", "×2"]);
  });

  it("draws ONE card when every member is in the same state", () => {
    const { container } = mountRow([soldier("s1"), soldier("s2"), soldier("s3")]);
    expect(drawnCards(container)).toEqual(["s1"]);
    expect(badges(container)).toEqual(["×3"]);
  });

  it("does not group a single token or printed cards", () => {
    const bear = (id: string): CardView => ({
      instance_id: id,
      name: "Grizzly Bears",
      owner: "me",
      controller: "me",
      type_line: "Creature — Bear",
      battle_x: Number(id.slice(1)),
    });
    const { container } = mountRow([soldier("s1"), bear("b2"), bear("b3")]);
    expect(drawnCards(container)).toEqual(["s1", "b2", "b3"]);
    expect(badges(container)).toEqual([]);
  });

  it("keeps an anchor for every hidden member on the group's card", () => {
    const { container } = mountRow([soldier("s1"), soldier("s2"), soldier("s3")]);
    for (const id of ["s1", "s2", "s3"]) {
      expect(container.querySelector(`[data-instance-id="${id}"]`)).not.toBeNull();
    }
  });

  it("rings the group's card when a hidden member is a legal target", () => {
    targeting.set({
      card: soldier("bolt", { name: "Lightning Bolt", type_line: "Instant" }),
      mode: "any",
      legal: { players: new Set(), cards: new Set(["s3"]) },
      min: 1,
      max: 1,
      picked: [],
      steps: [],
      step: 0,
      done: [],
    });
    const { container } = mountRow([soldier("s1"), soldier("s2"), soldier("s3")]);
    const drawn = container.querySelector<HTMLElement>('.card[data-instance-id="s1"]')!;
    expect(drawn.classList.contains("targetable")).toBe(true);
  });

  it("opens the list on a click instead of acting on one token", () => {
    const { container, opened, clicked } = mountRow([soldier("s1"), soldier("s2")]);
    click(container.querySelector(".card")!);
    expect(opened).toHaveLength(1);
    expect(clicked).toEqual([]);
  });

  it("still draws an Equipment on a grouped token behind the group's card", () => {
    const sword: CardView = {
      instance_id: "sword",
      name: "Bonesplitter",
      owner: "me",
      controller: "me",
      type_line: "Artifact — Equipment",
      attached_to: { kind: "card", id: "s2" },
    };
    const { container } = mountRow([soldier("s1"), soldier("s2", { power: 3 }), soldier("s3")], {
      s2: [sword],
    });
    expect(badges(container)).toEqual(["×3"]);
    expect(container.querySelector('.attachment [data-instance-id="sword"]')).not.toBeNull();
  });
});

function mountList(
  members: CardView[],
  opts: {
    combatMode?: "idle" | "attack" | "block";
    attachmentsByHost?: Record<string, CardView[]>;
  } = {},
) {
  const attacks: Array<{ ids: string[]; seat: string }> = [];
  const targeted: string[] = [];
  const used: string[] = [];
  let closed = 0;
  render(
    TokenGroupModal as never,
    {
      groupKey: "g",
      members,
      attachmentsByHost: opts.attachmentsByHost ?? {},
      view: snap(members),
      viewerID: "me",
      combatMode: opts.combatMode ?? "attack",
      canTap: true,
      onUse: (c: CardView) => used.push(c.instance_id),
      onTarget: (c: CardView) => {
        targeted.push(c.instance_id);
        return true;
      },
      onTapToggle: () => {},
      onAttack: (ids: string[], seatID: string) => attacks.push({ ids, seat: seatID }),
      onBlock: () => {},
      onClose: () => closed++,
    } as never,
  );
  return { attacks, targeted, used, closed: () => closed };
}

const body = () => document.body;
const rowButtons = () => [...body().querySelectorAll<HTMLButtonElement>('[role="checkbox"]')];
const checkedIDs = () =>
  rowButtons()
    .filter((b) => b.getAttribute("aria-checked") === "true")
    .map((b) => b.dataset.instance);
const button = (label: RegExp) =>
  [...body().querySelectorAll<HTMLButtonElement>("button")].find((b) => label.test(text(b)))!;
function setCount(n: number): void {
  const input = body().querySelector<HTMLInputElement>('input[type="number"]')!;
  input.value = String(n);
  input.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

describe("the token group's list (#1724)", () => {
  const members = () => [
    soldier("s1", { summoning_sick: true }),
    soldier("s2"),
    soldier("s3", { tapped: true }),
    soldier("s4"),
    soldier("s5"),
  ];

  it("registers a modal layer, so the global shortcuts stand down", () => {
    mountList(members());
    expect(get(modalDepth)).toBeGreaterThan(0);
  });

  it("attacks with ONE picked token", () => {
    const { attacks } = mountList(members());
    click(rowButtons()[3]);
    expect(checkedIDs()).toEqual(["s4"]);
    click(button(/^Attack Bob with 1$/));
    expect(attacks).toEqual([{ ids: ["s4"], seat: "bob" }]);
  });

  it("select N takes untapped, non-summoning-sick tokens first, and shows them before the attack", () => {
    const { attacks } = mountList(members());
    setCount(2);
    click(button(/^Select 2$/));
    expect(checkedIDs()).toEqual(["s2", "s4"]);
    expect(text(body().querySelector('[data-testid="tg-count"]'))).toBe("2 selected");
    click(button(/^Attack Bob with 2$/));
    expect(attacks).toEqual([{ ids: ["s2", "s4"], seat: "bob" }]);
  });

  it("all selects every token that can attack", () => {
    const { attacks } = mountList(members());
    click(button(/^All 3$/));
    expect(checkedIDs()).toEqual(["s2", "s4", "s5"]);
    click(button(/^Attack Bob with 3$/));
    expect(attacks).toEqual([{ ids: ["s2", "s4", "s5"], seat: "bob" }]);
  });

  it("shows a member's counters and attachments as badges on its row", () => {
    const sword: CardView = {
      instance_id: "sword",
      name: "Bonesplitter",
      owner: "me",
      controller: "me",
      type_line: "Artifact — Equipment",
    };
    mountList(
      [soldier("s1"), soldier("s2", { power: 2, toughness: 2, counters: { "+1/+1": 1 } })],
      { attachmentsByHost: { s1: [sword] } },
    );
    const [r1, r2] = rowButtons();
    expect(text(r1.querySelector(".tg-badges"))).toBe("Bonesplitter");
    expect(text(r2.querySelector(".tg-badges"))).toBe("+1/+1");
    expect(text(r2.querySelector(".pt"))).toBe("2/2");
  });

  it("a targeting prompt can pick one specific token from the list", () => {
    const t: TargetingState = {
      card: soldier("bolt", { name: "Lightning Bolt", type_line: "Instant" }),
      mode: "any",
      legal: { players: new Set(), cards: new Set(["s2", "s3", "s4"]) },
      min: 1,
      max: 1,
      picked: [],
      steps: [],
      step: 0,
      done: [],
    };
    targeting.set(t);
    const { targeted, closed } = mountList(members(), { combatMode: "idle" });
    // The illegal rows say so; the legal ones don't.
    expect(text(rowButtons()[0])).toContain("not a legal target");
    expect(text(rowButtons()[2])).not.toContain("not a legal target");
    click(rowButtons()[2]);
    click(button(/^Target 1$/));
    expect(targeted).toEqual(["s3"]);
    expect(closed()).toBe(1);
  });

  it("'Use this one' is the board click on that token", () => {
    const { used } = mountList(members());
    click(rowButtons()[4]);
    click(button(/^Use this one$/));
    expect(used).toEqual(["s5"]);
  });
});

// The panel end to end: the group's card opens the list, and what the
// list selects reaches the same actions a board click sends.
function mountPanel(cards: CardView[], combatMode: "idle" | "attack" | "block", view?: GameView) {
  const sent: Array<{ type: string; params?: unknown }> = [];
  const declared: Array<{ ids: string[]; seat: string }> = [];
  const tapped: string[] = [];
  const activated: Array<[string, number]> = [];
  const v = view ?? snap(cards);
  const r = render(
    PlayerPanel as never,
    {
      seat: v.seats[0],
      isSelf: true,
      isActive: true,
      hasPriority: true,
      viewerID: "me",
      isAdmin: false,
      sendAction: (type: string, params?: unknown) => sent.push({ type, params }),
      isMonarch: false,
      isInitiative: false,
      view: v,
      controlledCards: cards.filter((c) => c.controller === "me"),
      exile: zone("exile", undefined, []),
      combatMode,
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: (c: CardView) => tapped.push(c.instance_id),
      onPlayCard: () => {},
      onDrawCard: () => {},
      onManaAbilityCost: () => {},
      onActivateAbility: (c: CardView, i: number) => activated.push([c.instance_id, i]),
      onDeclareAttackers: (ids: string[], seatID: string) => declared.push({ ids, seat: seatID }),
    } as never,
  );
  const groupCard = () => r.container.querySelector<HTMLElement>(".pile.group .card")!;
  return { ...r, sent, declared, tapped, activated, groupCard };
}

describe("a token group in the player's panel (#1724)", () => {
  it("clicking the group opens the list; 'all' then attack declares exactly those tokens", () => {
    const cards = [soldier("s1", { summoning_sick: true }), soldier("s2"), soldier("s3")];
    const p = mountPanel(cards, "attack");
    expect(rowButtons()).toHaveLength(0);
    click(p.groupCard());
    expect(rowButtons()).toHaveLength(3);
    click(button(/^All 2$/));
    click(button(/^Attack Bob with 2$/));
    expect(p.declared).toEqual([{ ids: ["s2", "s3"], seat: "bob" }]);
    // The list closes once the declaration is sent.
    expect(rowButtons()).toHaveLength(0);
  });

  it("blocks with the selected tokens: one declare_blocker each", () => {
    const attacker: CardView = {
      instance_id: "ogre",
      name: "Ogre",
      owner: "bob",
      controller: "bob",
      type_line: "Creature — Ogre",
      power: 3,
      toughness: 3,
      attacking_target: "me",
      tapped: true,
    };
    const cards = [soldier("s1"), soldier("s2"), soldier("s3")];
    const p = mountPanel(cards, "block", snap([...cards, attacker], "declare_blockers"));
    click(p.groupCard());
    setCount(2);
    click(button(/^Select 2$/));
    click(button(/^Block with 2$/));
    expect(p.sent).toEqual([
      { type: "declare_blocker", params: { blocker: "s1", attacker: "ogre" } },
      { type: "declare_blocker", params: { blocker: "s2", attacker: "ogre" } },
    ]);
  });

  // ADR 0117 §1: "Use this one" is the board click on that member, so
  // it follows the click rule. A token with nothing usable does nothing.
  it("'Use this one' on a vanilla token does nothing, as its board click does", () => {
    const cards = [soldier("s1"), soldier("s2")];
    const p = mountPanel(cards, "idle", snap(cards, "main"));
    click(p.groupCard());
    click(rowButtons()[1]);
    click(button(/^Use this one$/));
    expect(p.sent).toEqual([]);
    expect(p.tapped).toEqual([]);
    expect(get(abilityPopover)).toBeNull();
  });

  // ADR 0117 as amended by #2201: a token with exactly one usable
  // ability activates it on the member the player chose, as its board
  // click would, and opens no popover.
  const clueSac = {
    index: 0,
    ref: "own:0",
    label: "{2}, Sacrifice this artifact: Draw a card.",
    mana_cost: "{2}",
  };
  const clue = (id: string, extra: Partial<CardView> = {}): CardView => ({
    instance_id: id,
    name: "Clue",
    owner: "me",
    controller: "me",
    type_line: "Token Artifact — Clue",
    battle_x: Number(id.slice(1)),
    activated_abilities: [clueSac],
    ...extra,
  });

  it("'Use this one' on a token with one usable ability activates it on that member", () => {
    const cards = [clue("c1"), clue("c2"), clue("c3")];
    const p = mountPanel(cards, "idle", snap(cards, "main"));
    click(p.groupCard());
    click(rowButtons()[1]);
    click(button(/^Use this one$/));
    expect(p.activated).toEqual([["c2", 0]]);
    expect(get(abilityPopover)).toBeNull();
    expect(p.tapped).toEqual([]);
  });

  // A token with two usable abilities opens its popover, at the
  // group's drawn card, on the member the player chose: the group
  // draws that member while its popover is open, and its rows act on
  // it.
  it("'Use this one' on a token with two usable abilities opens its popover at the group's card", () => {
    const twoRows = {
      activated_abilities: [
        clueSac,
        { index: 1, ref: "own:1", label: "{1}: Scry 1.", mana_cost: "{1}" },
      ],
    };
    const cards = [clue("c1", twoRows), clue("c2", twoRows), clue("c3", twoRows)];
    const p = mountPanel(cards, "idle", snap(cards, "main"));
    click(p.groupCard());
    click(rowButtons()[1]);
    click(button(/^Use this one$/));
    expect(get(abilityPopover)?.cardID).toBe("c2");
    // One group card, now drawn as the chosen member, with the popover.
    expect(p.container.querySelectorAll(".pile.group .card")).toHaveLength(1);
    expect(p.groupCard().dataset.instanceId).toBe("c2");
    const rows = [
      ...p
        .groupCard()
        .querySelectorAll<HTMLButtonElement>(".mana-menu .menu-item:not([data-raw-tap])"),
    ];
    expect(rows).toHaveLength(2);
    rows[0].click();
    flushSync();
    expect(p.activated).toEqual([["c2", 0]]);
    // Closed, the group draws its first member again.
    expect(get(abilityPopover)).toBeNull();
    expect(p.groupCard().dataset.instanceId).toBe("c1");
  });

  it("a Treasure picked from the list is clicked for mana, as it would be on the board", () => {
    const treasure = (id: string): CardView => ({
      instance_id: id,
      name: "Treasure",
      owner: "me",
      controller: "me",
      type_line: "Token Artifact — Treasure",
      battle_x: Number(id.slice(1)),
      mana_abilities: [
        { index: 0, tap_cost: true, sacrifice_cost: true, produced: "{C}", label: "Add {C}" },
      ],
    });
    const cards = [treasure("t1"), treasure("t2")];
    const p = mountPanel(cards, "idle", snap(cards, "main"));
    click(p.groupCard());
    click(rowButtons()[1]);
    click(button(/^Use this one$/));
    expect(p.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: expect.objectContaining({ card_id: "t2", ability_index: 0 }),
      },
    ]);
  });
});
