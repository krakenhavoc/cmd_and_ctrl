// @vitest-environment jsdom
//
// #1865 (ADR 0106 §1 amendment 2026-10-16): a spell whose ability works
// on the stack (Lightning Storm) carries it on its stack card's
// `zone_abilities`, which the server sends only to a seat that may
// activate it. The lane model lifts the rows onto the item, every stack
// surface draws a button for each beside Counter, and a click hands the
// row's wire index to the announce chain. The spell's counters are a
// chip.

import { describe, it, expect, afterEach } from "vitest";

import StackLanePile from "./components/board/StackLanePile.svelte";
import StackOverlay from "./components/board/StackOverlay.svelte";
import { buildStackLane, type StackLaneControls } from "./stackLane";
import type { CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { L } from "./labels";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(() => cleanup());

const ME = "me";
const OPP = "opp";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, name: string, n: number): PlayerView =>
  ({
    id,
    name,
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

const STORM_ROW = {
  index: 0,
  ref: "own:0",
  label:
    "Discard a land card: Put two charge counters on Lightning Storm. You may choose a new target for it. Any player may activate this ability but only if Lightning Storm is on the stack.",
  any_player: true,
  discard_cost_n: 1,
  discard_cost_label: "a land card",
};

function storm(withRows: boolean, counters?: Record<string, number>): CardView {
  return {
    instance_id: "storm",
    name: "Lightning Storm",
    owner: OPP,
    controller: OPP,
    known_by_you: true,
    type_line: "Instant",
    counters,
    zone_abilities: withRows ? [STORM_ROW] : undefined,
  } as unknown as CardView;
}

const stormItem: StackItemView = {
  id: "storm",
  kind: "spell",
  controller: OPP,
  owner: OPP,
  source_card_id: "storm",
  targets: [{ kind: "player", id: ME }],
} as StackItemView;

function view(card: CardView): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Bot 2", 1)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined, [card]),
    exile: zone("exile", undefined),
    stack_items: [stormItem],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 1, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  } as unknown as GameView;
}

function model(v: GameView, viewerHolds = true) {
  return buildStackLane({
    stack: v.stack,
    stackItems: v.stack_items,
    pendingTriggers: v.pending_triggers,
    seats: v.seats,
    battlefield: v.battlefield,
    exile: v.exile,
    viewerID: ME,
    priorityHolder: viewerHolds ? ME : OPP,
    splitSecondActive: false,
  });
}

function controls(withActivate = true) {
  const activated: [string, number][] = [];
  const c: StackLaneControls = {
    counter: () => {},
    targetable: () => false,
    target: () => {},
    hoverEnter: () => {},
    hoverLeave: () => {},
    activate: withActivate ? (it, i) => activated.push([it.id, i]) : undefined,
  };
  return { c, activated };
}

const NAME = L.stackAbility("Lightning Storm", "Discard a land card");

describe("a spell's ability on the stack (#1865)", () => {
  it("lifts the stack card's rows onto the item, and only a spell's", () => {
    const m = model(view(storm(true)));
    expect(m.stackItems[0].abilities.map((a) => a.ref)).toEqual(["own:0"]);
    expect(model(view(storm(false))).stackItems[0].abilities).toEqual([]);
  });

  it("shows the spell's counters as a chip", () => {
    const m = model(view(storm(true, { charge: 2 })));
    expect(m.stackItems[0].chips.map((c) => c.label)).toContain("2 charge");
  });

  it("draws a button beside Counter that activates the row by its index", () => {
    const { c, activated } = controls();
    const r = render(
      StackLanePile as never,
      {
        model: model(view(storm(true))),
        controls: c,
        styleName: "pile",
        pile: { cardWidth: 270, shrunk: false },
      } as never,
    );
    const btn = r.container.querySelector<HTMLButtonElement>(`button[aria-label="${NAME}"]`);
    expect(btn).not.toBeNull();
    expect(btn!.disabled).toBe(false);
    expect(btn!.textContent?.trim()).toBe("Discard a land card");
    click(btn!);
    expect(activated).toEqual([["storm", 0]]);
  });

  it("greys the button while the viewer does not hold priority", () => {
    const r = render(
      StackLanePile as never,
      {
        model: model(view(storm(true)), false),
        controls: controls().c,
        styleName: "pile",
        pile: { cardWidth: 270, shrunk: false },
      } as never,
    );
    const btn = r.container.querySelector<HTMLButtonElement>(`button[aria-label="${NAME}"]`);
    expect(btn?.disabled).toBe(true);
    expect(btn?.title).toContain("you don't hold priority");
  });

  it("draws nothing with no rows, or with nothing to announce through", () => {
    const none = render(
      StackLanePile as never,
      {
        model: model(view(storm(false))),
        controls: controls().c,
        styleName: "pile",
        pile: { cardWidth: 270, shrunk: false },
      } as never,
    );
    expect(none.container.querySelector(`button[aria-label="${NAME}"]`)).toBeNull();
    cleanup();
    const spectator = render(
      StackLanePile as never,
      {
        model: model(view(storm(true))),
        controls: controls(false).c,
        styleName: "pile",
        pile: { cardWidth: 270, shrunk: false },
      } as never,
    );
    expect(spectator.container.querySelector(`button[aria-label="${NAME}"]`)).toBeNull();
  });

  it("the docked card hands the spell's stack card to the announce chain", () => {
    const v = view(storm(true));
    const calls: [string, number][] = [];
    const r = render(
      StackOverlay as never,
      {
        stack: v.stack,
        battlefield: v.battlefield,
        exile: v.exile,
        stackItems: v.stack_items,
        pendingTriggers: [],
        seats: v.seats,
        viewerHasPriority: true,
        splitSecondActive: false,
        onCounter: () => {},
        onActivateAbility: (card: CardView, i: number) => calls.push([card.instance_id, i]),
      } as never,
    );
    const btn = r.container.querySelector<HTMLButtonElement>(`button[aria-label="${NAME}"]`);
    expect(btn).not.toBeNull();
    click(btn!);
    expect(calls).toEqual([["storm", 0]]);
  });

  // #4043: the storm is not on top; a response sits above it.
  function underResponse(): GameView {
    const v = view(storm(true));
    const resp = {
      instance_id: "resp",
      name: "Shock",
      owner: ME,
      controller: ME,
      known_by_you: true,
      type_line: "Instant",
    } as unknown as CardView;
    v.stack = zone("stack", undefined, [storm(true), resp]);
    v.stack_items = [
      stormItem,
      {
        id: "resp",
        kind: "spell",
        controller: ME,
        owner: ME,
        source_card_id: "resp",
        targets: [],
      } as StackItemView,
    ];
    return v;
  }

  it("offers the ability on a card under a response, on its strip", () => {
    const { c, activated } = controls();
    const r = render(
      StackLanePile as never,
      {
        model: model(underResponse()),
        controls: c,
        styleName: "pile",
        pile: { cardWidth: 270, shrunk: false },
      } as never,
    );
    const btn = r.container.querySelector<HTMLButtonElement>(`.peek button[aria-label="${NAME}"]`);
    expect(btn).not.toBeNull();
    expect(btn!.disabled).toBe(false);
    // The strip is slim: the button is an icon, the name and tooltip carry the cost.
    expect(btn!.textContent?.trim()).toBe("");
    expect(btn!.title).toContain("Discard a land card");
    click(btn!);
    expect(activated).toEqual([["storm", 0]]);
    // The top card (the response) has none.
    expect(r.container.querySelectorAll(`button[aria-label="${NAME}"]`).length).toBe(1);
  });

  it("keeps the caster's name on a row of its own, above the top card's ability", () => {
    const r = render(
      StackLanePile as never,
      {
        model: model(view(storm(true))),
        controls: controls().c,
        styleName: "pile",
        pile: { cardWidth: 270, shrunk: false },
      } as never,
    );
    const btn = r.container.querySelector(`button[aria-label="${NAME}"]`)!;
    expect(btn.closest(".who-row")).toBeNull();
    expect(btn.closest(".ability-row")).not.toBeNull();
  });
});
