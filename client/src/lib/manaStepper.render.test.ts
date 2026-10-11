// @vitest-environment jsdom
//
// ADR 0117 §4 on the real Board: the per-colour stepper. A left-click
// on a source whose one usable ability adds two or more mana with a
// colour choice (Vivi Ornitier, untapped or tapped) opens the anchored
// picker straight on its stepper; the light popover's mana row (a
// right-click) opens the same stepper instead of activating with no
// colours; a picker listing several abilities opens it in place. Add
// mana sends ONE activate_mana_ability carrying one colour per slot.
//
// The accessible names asserted here are the ADR's table, and a
// contract (AGENTS.md §5): the dialog "Split N mana from <card>", one
// group per colour ("blue"), its "less blue" / "more blue" buttons and
// "blue count", the "N of N" status and the "Add mana" button.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";
import { tick } from "svelte";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import { closeManaSourcePicker, manaSourcePicker } from "./manaSourcePicker";
import { forgetSplits } from "./manaStepper";
import type { ActionType, CardView, GameView, ManaAbilityView, PlayerView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

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
  forgetSplits();
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

// Vivi Ornitier as the server ships her: "{0}" (no {T}), one {U|R}
// slot per point of power, a list per slot.
const viviMana = (power: number): ManaAbilityView => ({
  index: 0,
  ref: "own:m0",
  mana_cost: "{0}",
  label:
    "{0}: Add X mana in any combination of {U} and/or {R}, where X is Vivi Ornitier's power. Activate only during your turn and only once each turn.",
  produced: Array.from({ length: power }, () => "{U|R}").join(""),
  color_options: Array.from({ length: power }, () => ["U", "R"]),
});

const vivi = (power: number, extra: Partial<CardView> = {}) =>
  permanent("vivi", "Vivi Ornitier", "Legendary Creature — Wizard", {
    power,
    toughness: 3,
    mana_abilities: [viviMana(power)],
    ...extra,
  });

const gate = () =>
  permanent("gate", "Mystic Gate", "Land", {
    mana_abilities: [
      { index: 0, ref: "own:m0", tap_cost: true, produced: "{C}", label: "{T}: Add {C}." },
      {
        index: 1,
        ref: "own:m1",
        tap_cost: true,
        mana_cost: "{W/U}",
        produced: "{W|U}{W|U}",
        label: "{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}{U}.",
        color_options: [
          ["W", "U"],
          ["W", "U"],
        ],
      },
    ],
  });

// A fixed slot beside a wide one: Command Tower narrowed to green (CR
// 903.4f) shape, as a test of the fixed row. No catalog card has it.
const fixedBeside = () =>
  permanent("odd", "Odd Source", "Artifact", {
    mana_abilities: [
      {
        index: 0,
        ref: "own:m0",
        tap_cost: true,
        produced: "{W|U|B|R|G}{W|U|B|R|G}",
        color_options: [["G"], ["G", "W", "U", "B", "R"]],
      },
    ],
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
  const dialog = () => document.querySelector<HTMLElement>(".mana-source-picker");
  const named = <T extends HTMLElement = HTMLElement>(label: string) =>
    dialog()?.querySelector<T>(`[aria-label="${label}"]`) ?? null;
  const button = (label: string) => named<HTMLButtonElement>(label)!;
  const count = (color: string) => named(`${color} count`)?.textContent?.trim();
  const status = () => dialog()?.querySelector<HTMLElement>("[role=status]") ?? null;
  const addMana = () =>
    Array.from(dialog()?.querySelectorAll<HTMLButtonElement>("button") ?? []).find(
      (b) => b.textContent?.trim() === "Add mana",
    )!;
  const groups = () =>
    Array.from(dialog()?.querySelectorAll<HTMLElement>(".split [role=group]") ?? []).map((g) =>
      g.getAttribute("aria-label"),
    );
  const rightClick = (id: string) => {
    tile(id).dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    flushSync();
  };
  // #2960: the popover is drawn in the board's popover host, not in the
  // card, and only one is ever open. `id` is kept for the reader.
  const popoverManaRow = (_id: string) =>
    document.querySelector<HTMLButtonElement>(
      "[data-popover-host] .mana-menu .menu-item[data-kind='mana']",
    )!;
  const setView = (cards: CardView[]) => r.setProps({ view: gameView(cards) } as never);
  return {
    ...r,
    sent,
    tile,
    dialog,
    button,
    count,
    status,
    addMana,
    groups,
    rightClick,
    popoverManaRow,
    setView,
  };
}

const key = (k: string) => {
  window.dispatchEvent(new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true }));
  flushSync();
};

describe("the stepper's names and counts (ADR 0117 §4)", () => {
  it("opens on a tapped Vivi at power 3 and sends one activation with three colours, no untap", () => {
    const b = mountBoard([vivi(3, { tapped: true })]);
    click(b.tile("vivi"));
    expect(get(manaSourcePicker)).toMatchObject({ cardID: "vivi", abilityIndex: 0 });
    expect(b.sent).toEqual([]);

    const d = b.dialog()!;
    expect(d.getAttribute("role")).toBe("dialog");
    expect(d.getAttribute("aria-label")).toBe("Split 3 mana from Vivi Ornitier");
    // One group per colour, in the server's order.
    expect(b.groups()).toEqual(["blue", "red"]);
    for (const c of ["blue", "red"]) {
      expect(b.button(`less ${c}`).tagName).toBe("BUTTON");
      expect(b.button(`more ${c}`).tagName).toBe("BUTTON");
    }
    // The start state: all on the first colour every list offers.
    expect(b.count("blue")).toBe("3");
    expect(b.count("red")).toBe("0");
    const status = b.status()!;
    expect(status.getAttribute("aria-live")).toBe("polite");
    expect(status.getAttribute("aria-label")).toBe("3 of 3");
    expect(status.textContent?.trim()).toBe("3 of 3");
    expect(b.addMana().disabled).toBe(false);

    // Short of the total: Add mana waits.
    click(b.button("less blue"));
    expect(b.status()!.textContent?.trim()).toBe("2 of 3");
    expect(b.addMana().disabled).toBe(true);
    // + on either colour is live while short; − on red is not at 0.
    expect(b.button("more red").disabled).toBe(false);
    click(b.button("less blue"));
    expect(b.button("less red").disabled).toBe(true);
    click(b.button("more red"));
    click(b.button("more red"));
    // At the total, every + is off.
    expect(b.status()!.textContent?.trim()).toBe("3 of 3");
    expect(b.button("more red").disabled).toBe(true);
    expect(b.button("more blue").disabled).toBe(true);
    expect(b.count("blue")).toBe("1");
    expect(b.count("red")).toBe("2");

    click(b.addMana());
    expect(b.sent).toHaveLength(1);
    const [only] = b.sent;
    expect(only.type).toBe("activate_mana_ability");
    expect(only.params).toMatchObject({ card_id: "vivi", ability_index: 0 });
    const colors = only.params!.colors as string[];
    expect(colors).toHaveLength(3);
    expect([...colors].sort()).toEqual(["R", "R", "U"]);
    expect(only.params!.color).toBeUndefined();
    expect(b.sent.some((s) => s.type === "untap" || s.type === "tap")).toBe(false);
    expect(get(manaSourcePicker)).toBeNull();
  });

  it("keeps a Vivi at power 1 on her two buttons", () => {
    const b = mountBoard([vivi(1)]);
    click(b.tile("vivi"));
    expect(b.dialog()?.getAttribute("aria-label")).toBe("Tap Vivi Ornitier for mana");
    expect(b.dialog()?.querySelector(".split")).toBeNull();
    expect(b.dialog()?.querySelectorAll(".mana-option")).toHaveLength(2);
  });

  it("starts from the split last confirmed for this card", () => {
    const b = mountBoard([vivi(3)]);
    click(b.tile("vivi"));
    click(b.button("less blue"));
    click(b.button("more red"));
    click(b.addMana());
    expect(b.sent).toHaveLength(1);
    click(b.tile("vivi"));
    expect(b.count("blue")).toBe("2");
    expect(b.count("red")).toBe("1");
  });

  it("re-reads the lists each frame and resets when their number changes", () => {
    const b = mountBoard([vivi(3)]);
    click(b.tile("vivi"));
    click(b.button("less blue"));
    expect(b.status()!.textContent?.trim()).toBe("2 of 3");
    // Vivi's power went up while the stepper was open.
    b.setView([vivi(4)]);
    expect(b.dialog()?.getAttribute("aria-label")).toBe("Split 4 mana from Vivi Ornitier");
    expect(b.status()!.textContent?.trim()).toBe("4 of 4");
    expect(b.count("blue")).toBe("4");
    // An ordinary frame with the same lists keeps the player's counts.
    click(b.button("less blue"));
    b.setView([vivi(4)]);
    expect(b.count("blue")).toBe("3");
  });

  it("shows a fixed slot as a count that cannot be stepped away", () => {
    const b = mountBoard([fixedBeside()]);
    click(b.tile("odd"));
    expect(b.dialog()?.getAttribute("aria-label")).toBe("Split 2 mana from Odd Source");
    expect(b.groups()).toEqual(["green", "white", "blue", "black", "red"]);
    // Start: both on green, the one colour every list offers.
    expect(b.count("green")).toBe("2");
    click(b.button("less green"));
    expect(b.count("green")).toBe("1");
    // Green's last one is the fixed slot's share.
    expect(b.button("less green").disabled).toBe(true);
    click(b.button("more red"));
    // Full: red cannot take the fixed slot.
    expect(b.button("more red").disabled).toBe(true);
    click(b.addMana());
    expect(b.sent[0].params!.colors).toEqual(["G", "R"]);
  });
});

describe("the stepper's keyboard", () => {
  it("Enter confirms; Up/Down move between rows; Left/Right and −/+ step", async () => {
    const b = mountBoard([vivi(3)]);
    click(b.tile("vivi"));
    await tick();
    // Down to red, + twice, then back up to blue and − twice.
    key("ArrowDown");
    key("ArrowRight");
    expect(b.count("red")).toBe("0"); // the total is met: + does nothing
    key("ArrowUp");
    key("ArrowLeft");
    key("-");
    expect(b.count("blue")).toBe("1");
    key("ArrowDown");
    key("+");
    key("ArrowRight");
    expect(b.count("red")).toBe("2");
    key("Enter");
    expect(b.sent).toHaveLength(1);
    expect([...(b.sent[0].params!.colors as string[])].sort()).toEqual(["R", "R", "U"]);
  });

  it("Enter does nothing while the total is short", () => {
    const b = mountBoard([vivi(3)]);
    click(b.tile("vivi"));
    click(b.button("less blue"));
    key("Enter");
    expect(b.sent).toEqual([]);
    expect(b.dialog()).not.toBeNull();
  });

  it("Escape cancels and sends nothing", () => {
    const b = mountBoard([vivi(3, { tapped: true })]);
    click(b.tile("vivi"));
    expect(b.dialog()).not.toBeNull();
    key("Escape");
    expect(b.dialog()).toBeNull();
    expect(get(manaSourcePicker)).toBeNull();
    expect(b.sent).toEqual([]);
  });
});

describe("other routes to the stepper", () => {
  it("a right-click on Vivi opens the stepper, not an activation with no colours", () => {
    const b = mountBoard([vivi(3)]);
    b.rightClick("vivi");
    click(b.popoverManaRow("vivi"));
    expect(b.sent).toEqual([]);
    expect(b.dialog()?.getAttribute("aria-label")).toBe("Split 3 mana from Vivi Ornitier");
    click(b.addMana());
    expect(b.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "vivi", ability_index: 0, ref: "own:m0", colors: ["U", "U", "U"] },
      },
    ]);
  });

  it("a popover mana row with one picking slot opens its colour buttons", () => {
    const birds = permanent("birds", "Birds of Paradise", "Creature — Bird", {
      mana_abilities: [
        {
          index: 0,
          ref: "own:m0",
          tap_cost: true,
          produced: "{W|U|B|R|G}",
          color_options: [["G", "W", "U", "B", "R"]],
        },
      ],
    });
    const b = mountBoard([birds]);
    b.rightClick("birds");
    click(b.popoverManaRow("birds"));
    expect(b.sent).toEqual([]);
    expect(b.dialog()?.getAttribute("aria-label")).toBe("Tap Birds of Paradise for mana");
    const opts = b.dialog()!.querySelectorAll<HTMLButtonElement>(".mana-option");
    expect(opts).toHaveLength(5);
    click(opts[1]);
    expect(b.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "birds", ability_index: 0, ref: "own:m0", color: "W" },
      },
    ]);
  });

  it("a popover mana row with no colour choice still activates at once", () => {
    const swamp = permanent("swamp", "Swamp", "Basic Land — Swamp", {
      mana_abilities: [{ index: 0, ref: "own:m0", tap_cost: true, produced: "{B}" }],
    });
    const b = mountBoard([swamp]);
    b.rightClick("swamp");
    click(b.popoverManaRow("swamp"));
    expect(b.dialog()).toBeNull();
    expect(b.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "swamp", ability_index: 0, ref: "own:m0" },
      },
    ]);
  });

  it("a picker of several abilities opens the filter row's stepper in place (Mystic Gate)", () => {
    const b = mountBoard([gate()]);
    click(b.tile("gate"));
    expect(b.dialog()?.getAttribute("aria-label")).toBe("Tap Mystic Gate for mana");
    const opts = b.dialog()!.querySelectorAll<HTMLButtonElement>(".mana-option");
    expect(opts).toHaveLength(2);
    expect(opts[1].textContent).toContain("2 mana, any split");
    click(opts[1]);
    expect(b.sent).toEqual([]);
    expect(b.dialog()?.getAttribute("aria-label")).toBe("Split 2 mana from Mystic Gate");
    expect(b.groups()).toEqual(["white", "blue"]);
    click(b.button("less white"));
    click(b.button("more blue"));
    click(b.addMana());
    expect(b.sent).toHaveLength(1);
    expect(b.sent[0].params).toMatchObject({ card_id: "gate", ability_index: 1 });
    expect([...(b.sent[0].params!.colors as string[])].sort()).toEqual(["U", "W"]);
  });
});
