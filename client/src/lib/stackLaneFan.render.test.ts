// @vitest-environment jsdom
//
// #1467 — design A of the floating stack lane, the fan. The model is
// pinned in stackLane.test.ts and the arrow geometry in
// stackArrows.test.ts; this file pins what exists only once the fan is
// rendered: tile order and sizes, the Next badge, the verbatim names
// e2e specs look for, that each button reaches its control, and that
// an arrow and a highlight ring appear for a permanent on the board.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import Board from "./components/board/Board.svelte";
import StackLaneFan from "./components/board/StackLaneFan.svelte";
import { updateSettings } from "./settings";
import {
  buildStackLane,
  type StackLaneControls,
  type StackLaneItem,
  type StackLaneModel,
} from "./stackLane";
import { TARGET_ATTR } from "./stackArrows";
import type { ActionType, CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

const ME = "me";
const OPP = "opp";

const realRect = Element.prototype.getBoundingClientRect;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  // jsdom does no layout, and an element with no size is never an
  // arrow end. Give every element the same box, except board cards,
  // which sit below it — a target under the lane gets no arrow.
  Element.prototype.getBoundingClientRect = function (this: Element) {
    const top = this.hasAttribute("data-instance-id") ? 900 : 10;
    return {
      left: 10,
      top,
      width: 100,
      height: 140,
      right: 110,
      bottom: top + 140,
      x: 10,
      y: top,
      toJSON: () => ({}),
    } as DOMRect;
  };
});

afterEach(() => {
  cleanup();
  Element.prototype.getBoundingClientRect = realRect;
  updateSettings("display", "stackStyle", "compact");
});

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

const card = (id: string, name: string, controller: string, extra: Partial<CardView> = {}) =>
  ({
    instance_id: id,
    name,
    owner: controller,
    controller,
    known_by_you: true,
    ...extra,
  }) as CardView;

const birds = card("birds", "Birds of Paradise", OPP, {
  type_line: "Creature — Bird",
  power: 0,
  toughness: 1,
});
const bolt = card("bolt", "Lightning Bolt", ME, { type_line: "Instant" });
const counterspell = card("cs", "Counterspell", OPP, { type_line: "Instant" });
const vivi = card("vivi", "Vivi Ornitier", ME, {
  type_line: "Legendary Creature — Wizard",
  power: 0,
  toughness: 3,
});

const spell = (id: string, controller: string, targets: StackItemView["targets"] = []) =>
  ({
    id,
    kind: "spell",
    controller,
    owner: controller,
    source_card_id: id,
    targets,
  }) as StackItemView;

// Bottom first, as the wire sends it: Vivi's trigger, then Lightning
// Bolt at Birds, then Counterspell at the Bolt — so the lane reads
// Counterspell, Lightning Bolt, Vivi's trigger.
const viviTrigger: StackItemView = {
  id: "trig-vivi",
  kind: "triggered",
  controller: ME,
  owner: ME,
  source_card_id: "vivi",
  label: "Vivi Ornitier — put a +1/+1 counter on Vivi",
};
const stackItems: StackItemView[] = [
  viviTrigger,
  spell("bolt", ME, [{ kind: "card", id: "birds" }]),
  spell("cs", OPP, [{ kind: "card", id: "bolt" }]),
];
const pendingTrigger: StackItemView = {
  id: "pend-1",
  kind: "triggered",
  controller: OPP,
  owner: OPP,
  source_card_id: "birds",
  label: "Birds of Paradise — waiting trigger",
};

const gameView = (items: StackItemView[], pending: StackItemView[] = []): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Bot 2", 1)],
    battlefield: zone("battlefield", undefined, [birds, vivi]),
    stack: zone("stack", undefined, [bolt, counterspell]),
    exile: zone("exile", undefined),
    stack_items: items,
    pending_triggers: pending,
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  }) as unknown as GameView;

function laneModel(pending: StackItemView[] = []): StackLaneModel {
  const v = gameView(stackItems, pending);
  return buildStackLane({
    stack: v.stack,
    stackItems: v.stack_items,
    pendingTriggers: v.pending_triggers,
    seats: v.seats,
    battlefield: v.battlefield,
    exile: v.exile,
    viewerID: ME,
    priorityHolder: ME,
    splitSecondActive: false,
  });
}

function fakeControls(targetable = false) {
  const calls = {
    counter: [] as string[],
    target: [] as string[],
    enter: [] as string[],
    leave: [] as string[],
  };
  const controls: StackLaneControls = {
    counter: (it: StackLaneItem) => calls.counter.push(it.id),
    pass: undefined,
    targetable: () => targetable,
    target: (it: StackLaneItem) => calls.target.push(it.id),
    hoverEnter: (it: StackLaneItem) => calls.enter.push(it.id),
    hoverLeave: (it: StackLaneItem) => calls.leave.push(it.id),
  };
  return { controls, calls };
}

function mountFan(opts: { pending?: StackItemView[]; targetable?: boolean } = {}) {
  const { controls, calls } = fakeControls(opts.targetable);
  const r = render(
    StackLaneFan as never,
    {
      model: laneModel(opts.pending),
      controls,
      styleName: "fan",
    } as never,
  );
  const qa = (sel: string) => Array.from(r.container.querySelectorAll<HTMLElement>(sel));
  return { ...r, calls, qa };
}

describe("the fan stack lane (#1467)", () => {
  it("lays the stack out top first, the top tile largest and marked next", () => {
    const f = mountFan();
    const tiles = f.qa("[data-stack-item-id]");
    expect(tiles.map((t) => t.dataset.stackItemId)).toEqual(["cs", "bolt", "trig-vivi"]);
    // Names verbatim: e2e specs look for them.
    expect(tiles.map((t) => t.querySelector(".name")?.textContent)).toEqual([
      "Counterspell",
      "Lightning Bolt",
      "Vivi Ornitier — put a +1/+1 counter on Vivi",
    ]);
    expect(f.qa(".slot").map((s) => s.dataset.size)).toEqual(["top", "second", "rest"]);
    const top = f.qa(".slot.top");
    expect(top).toHaveLength(1);
    expect(top[0].querySelector(".next-badge")?.textContent).toBe("Next");
    expect(f.qa(".next-badge")).toHaveLength(1);
  });

  it("tags an ability with its kind and captions every tile with caster and target", () => {
    const f = mountFan();
    const [cs, boltSlot, trig] = f.qa(".slot");
    expect(trig.querySelector(".tag")?.textContent).toBe("Trigger");
    expect(cs.querySelector(".tag")).toBeNull();
    expect(cs.querySelector(".caster")?.textContent).toBe("Bot 2");
    expect(cs.querySelector(".aim")?.textContent).toBe("→ your Lightning Bolt");
    expect(boltSlot.querySelector(".caster")?.textContent).toBe("You");
    expect(boltSlot.querySelector(".aim")?.textContent).toBe("→ Bot 2's Birds of Paradise");
  });

  it("rings a tile another item targets", () => {
    const f = mountFan();
    const boltTile = f.qa('[data-stack-item-id="bolt"]')[0];
    expect(boltTile.classList.contains("targeted")).toBe(true);
    expect(f.qa('[data-stack-item-id="cs"]')[0].classList.contains("targeted")).toBe(false);
  });

  it("puts pending triggers last, waiting, with no Counter", () => {
    const f = mountFan({ pending: [pendingTrigger] });
    const tiles = f.qa("[data-stack-item-id]").map((t) => t.dataset.stackItemId);
    expect(tiles[tiles.length - 1]).toBe("pend-1");
    const slot = f.qa(".slot.pending")[0];
    expect(slot.querySelector(".tag")?.textContent).toBe("Waiting");
    expect(slot.querySelector(".counter-btn")).toBeNull();
    expect(f.qa(".counter-btn")).toHaveLength(3);
  });

  it("sends Counter through controls.counter, from a labelled button", () => {
    const f = mountFan();
    const btn = f.qa(".counter-btn")[1];
    expect(btn.tagName).toBe("BUTTON");
    expect(btn.getAttribute("aria-label")).toBe("Counter Lightning Bolt");
    click(btn);
    expect(f.calls.counter).toEqual(["bolt"]);
  });

  it("offers a real Target button only while an item is targetable", () => {
    expect(mountFan().qa(".target-btn")).toHaveLength(0);
    cleanup();
    const f = mountFan({ targetable: true });
    const btns = f.qa(".target-btn");
    expect(btns).toHaveLength(3);
    expect(btns[0].getAttribute("aria-label")).toBe("Target Counterspell");
    click(btns[0]);
    expect(f.calls.target).toEqual(["cs"]);
  });

  it("drives the hover preview from pointer and focus", () => {
    const f = mountFan();
    const slot = f.qa(".slot")[0];
    slot.dispatchEvent(new Event("pointerenter"));
    slot.dispatchEvent(new Event("pointerleave"));
    slot.querySelector<HTMLElement>(".counter-btn")!.focus();
    expect(f.calls.enter).toEqual(["cs", "cs"]);
    expect(f.calls.leave).toEqual(["cs"]);
  });
});

describe("the fan's arrows, on a real board", () => {
  // An opponent's seat renders as a summary in this two-seat test
  // board, so the permanent target is one the viewer controls: the
  // Bolt is aimed at Vivi.
  const boardStack: StackItemView[] = [
    viviTrigger,
    spell("bolt", ME, [{ kind: "card", id: "vivi" }]),
    spell("cs", OPP, [{ kind: "card", id: "bolt" }]),
  ];

  function mountBoard() {
    updateSettings("display", "stackStyle", "fan");
    const sent: { type: ActionType; params?: unknown }[] = [];
    const r = render(
      Board as never,
      {
        view: gameView(boardStack),
        viewerID: ME,
        isAdmin: false,
        sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
        onPassPriority: () => {},
      } as never,
    );
    const q = (sel: string) => r.container.querySelector<HTMLElement>(sel);
    return { ...r, sent, q };
  }

  it("mounts the fan for the fan setting", () => {
    const b = mountBoard();
    expect(b.q('.stack-lane [data-stack-body="fan"]')).not.toBeNull();
  });

  it("draws an arrow to a permanent target whose card is on the board, and rings it", () => {
    const b = mountBoard();
    const viviCard = b.q('.board [data-instance-id="vivi"]');
    expect(viviCard).not.toBeNull();
    const arrow = b.q('.board-arrows path[data-arrow-target="vivi"]');
    expect(arrow).not.toBeNull();
    expect(arrow?.getAttribute("d")).toMatch(/^M [\d.-]+ [\d.-]+ Q /);
    expect(arrow?.getAttribute("marker-end")).toMatch(/^url\(#stack-fan-/);
    // Counterspell → Lightning Bolt is an arc inside the lane.
    expect(b.q('.lane-arcs path[data-arrow-target="bolt"]')).not.toBeNull();
    // The target is lit while the lane points at it…
    expect(viviCard?.hasAttribute(TARGET_ATTR)).toBe(true);
    // …and let go when the stack empties.
    b.setProps({ view: gameView([]) } as never);
    expect(b.q('.board [data-instance-id="vivi"]')?.hasAttribute(TARGET_ATTR)).toBe(false);
  });

  it("draws no arrow when the target has left the board", () => {
    const b = mountBoard();
    expect(b.q('.board-arrows path[data-arrow-target="vivi"]')).not.toBeNull();
    const v = gameView(boardStack);
    b.setProps({
      view: { ...v, battlefield: zone("battlefield", undefined, [birds]) },
    } as never);
    expect(b.q('.board-arrows path[data-arrow-target="vivi"]')).toBeNull();
  });

  it("Counter on a fan tile sends the same verb the docked card does", () => {
    const b = mountBoard();
    click(b.q('.stack-lane .counter-btn[aria-label="Counter Counterspell"]')!);
    expect(b.sent).toEqual([{ type: "counter_spell", params: { instance_id: "cs" } }]);
  });
});
