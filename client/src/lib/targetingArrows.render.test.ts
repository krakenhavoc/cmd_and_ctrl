// @vitest-environment jsdom
//
// ADR 0119 §4 on a real board. The decisions are pinned as pure
// functions in targetingArrows.test.ts and stackArrows.test.ts; this
// file pins what only exists once the board is rendered:
//
//   - the rings on what the stack targets, in every stack style
//     (StackTargetRings), compact included;
//   - while the viewer chooses targets (TargetingArrows): the source's
//     glow, an arrow per pick, the dock fallback for a source that is
//     not on screen, and the follow arrow for a mouse but not a touch.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { flushSync } from "svelte";

import Board from "./components/board/Board.svelte";
import { updateSettings } from "./settings";
import { targeting, type TargetingState } from "./targeting";
import { STACK_STYLES } from "./stackLane";
import { TARGET_ATTR, TARGET_RING_VAR } from "./stackArrows";
import { seatColor } from "./colors";
import type { ActionType, CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

const ME = "me";
const OPP = "opp";

const realRect = Element.prototype.getBoundingClientRect;
const g = globalThis as Record<string, unknown>;
const realElementFromPoint = document.elementFromPoint;

// jsdom does no layout, so every element the arrows read gets a place
// of its own: the board, the opponent's creature high up, the hand card
// low down, the opponent's header at the top, and the dock's question
// line bottom right. Everything else is one small box.
const PLACES: [string, [number, number, number, number]][] = [
  ['[data-instance-id="birds"]', [300, 120, 100, 140]],
  ['[data-instance-id="vivi"]', [500, 480, 100, 140]],
  ['[data-instance-id="shock"]', [640, 740, 100, 140]],
  ['[data-seat-id="opp"]', [900, 20, 160, 40]],
  [".dock-question", [1180, 830, 200, 36]],
];

function rect(left: number, top: number, width: number, height: number): DOMRect {
  return {
    left,
    top,
    width,
    height,
    right: left + width,
    bottom: top + height,
    x: left,
    y: top,
    toJSON: () => ({}),
  } as DOMRect;
}

beforeEach(() => {
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  // jsdom has no SVG geometry (CombatArrows' draw-on reads a path's
  // length for the docked styles) and no media playback (the hand's
  // card sounds).
  const svg = SVGElement.prototype as unknown as { getTotalLength?: () => number };
  svg.getTotalLength ??= () => 100;
  HTMLMediaElement.prototype.play = () => Promise.resolve();
  // The opponent's board drawn in full: a summarised seat draws pips,
  // not cards, so nothing of theirs could be ringed (ADR 0120 §1).
  updateSettings("display", "opponentDetail", "full");
  Element.prototype.getBoundingClientRect = function (this: Element) {
    if (this.classList.contains("board")) return rect(0, 0, 1400, 900);
    for (const [sel, r] of PLACES) if (this.matches(sel)) return rect(...r);
    return rect(10, 10, 100, 140);
  };
});

afterEach(() => {
  cleanup();
  Element.prototype.getBoundingClientRect = realRect;
  document.elementFromPoint = realElementFromPoint;
  targeting.set(null);
  updateSettings("display", "stackStyle", "compact");
  updateSettings("display", "opponentDetail", "summary");
  document.querySelectorAll("[data-test-dock]").forEach((el) => el.remove());
});

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const card = (id: string, name: string, controller: string, extra: Partial<CardView> = {}) =>
  ({
    instance_id: id,
    name,
    owner: controller,
    controller,
    known_by_you: true,
    scryfall_id: `sf-${id}`,
    ...extra,
  }) as CardView;

const birds = card("birds", "Birds of Paradise", OPP, { type_line: "Creature — Bird" });
const vivi = card("vivi", "Vivi Ornitier", OPP, { type_line: "Legendary Creature — Wizard" });
const shock = card("shock", "Shock", ME, { type_line: "Instant" });
const bolt = card("bolt", "Lightning Bolt", OPP, { type_line: "Instant" });
const growth = card("growth", "Giant Growth", ME, { type_line: "Instant" });

const seat = (id: string, name: string, n: number, hand: CardView[] = []): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const spell = (id: string, controller: string, targets: StackItemView["targets"] = []) =>
  ({
    id,
    kind: "spell",
    controller,
    owner: controller,
    source_card_id: id,
    targets,
  }) as StackItemView;

const gameView = (items: StackItemView[], stackCards: CardView[] = []): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0, [shock]), seat(OPP, "Bot 2", 1)],
    battlefield: zone("battlefield", undefined, [birds, vivi]),
    stack: zone("stack", undefined, stackCards),
    exile: zone("exile", undefined),
    stack_items: items,
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  }) as unknown as GameView;

function mountBoard(view: GameView) {
  const r = render(
    Board as never,
    {
      view,
      viewerID: ME,
      isAdmin: false,
      sendAction: (_type: ActionType, _params?: unknown) => {},
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
  const q = (sel: string) => r.container.querySelector<HTMLElement>(sel);
  const qa = (sel: string) => Array.from(r.container.querySelectorAll<HTMLElement>(sel));
  return { ...r, q, qa };
}

const nextFrame = () => new Promise((resolve) => setTimeout(resolve, 40));

function pointer(type: string, x: number, y: number): void {
  const e = new MouseEvent(type === "touch" ? "pointerdown" : "pointermove", {
    clientX: x,
    clientY: y,
    bubbles: true,
  });
  Object.defineProperty(e, "pointerType", { value: type });
  window.dispatchEvent(e);
}

function prompt(over: Partial<TargetingState> = {}): TargetingState {
  return {
    card: shock,
    mode: "any",
    legal: { players: new Set([OPP]), cards: new Set(["birds", "vivi"]) },
    min: 1,
    max: 2,
    picked: [],
    steps: [],
    step: 0,
    done: [],
    ...over,
  } as TargetingState;
}

describe("rings on what the stack targets (ADR 0119 §4)", () => {
  // Bottom first: Giant Growth (mine) aims at Birds, and Lightning Bolt
  // (the opponent's, on top) aims at Vivi and at me... and at Birds too,
  // so the top item's gold wins Birds' ring.
  const items = [
    spell("growth", ME, [{ kind: "card", id: "birds" }]),
    spell("bolt", OPP, [
      { kind: "card", id: "vivi" },
      { kind: "card", id: "birds" },
    ]),
  ];

  for (const style of STACK_STYLES) {
    it(`rings targeted permanents in the ${style} style, gold for the top item`, () => {
      updateSettings("display", "stackStyle", style);
      const b = mountBoard(gameView(items, [growth, bolt]));
      const viviEl = b.q('.board [data-instance-id="vivi"]')!;
      const birdsEl = b.q('.board [data-instance-id="birds"]')!;
      expect(viviEl.getAttribute(TARGET_ATTR)).toBe("next");
      expect(viviEl.style.getPropertyValue(TARGET_RING_VAR)).toBe("var(--gold)");
      expect(birdsEl.getAttribute(TARGET_ATTR)).toBe("next");

      // The top item resolves: Giant Growth is on top now, so Birds is
      // ringed gold by it and Vivi is not ringed at all.
      b.setProps({ view: gameView([items[0]], [growth]) } as never);
      expect(b.q('.board [data-instance-id="vivi"]')!.hasAttribute(TARGET_ATTR)).toBe(false);
      expect(b.q('.board [data-instance-id="birds"]')!.getAttribute(TARGET_ATTR)).toBe("next");

      b.setProps({ view: gameView([]) } as never);
      expect(b.qa(`[${TARGET_ATTR}]`)).toHaveLength(0);
    });
  }

  it("rings a lower item's target in its caster's colour", () => {
    updateSettings("display", "stackStyle", "compact");
    const b = mountBoard(
      gameView(
        [spell("growth", ME, [{ kind: "card", id: "birds" }]), spell("bolt", OPP)],
        [growth, bolt],
      ),
    );
    const birdsEl = b.q('.board [data-instance-id="birds"]')!;
    expect(birdsEl.getAttribute(TARGET_ATTR)).toBe("");
    expect(birdsEl.style.getPropertyValue(TARGET_RING_VAR)).toBe(seatColor(0));
  });

  it("rings a targeted player's header", () => {
    updateSettings("display", "stackStyle", "ribbon");
    const b = mountBoard(gameView([spell("bolt", OPP, [{ kind: "player", id: OPP }])], [bolt]));
    expect(b.q(`.board [data-seat-id="${OPP}"]`)?.getAttribute(TARGET_ATTR)).toBe("next");
  });

  it("clears every ring when the board goes", () => {
    updateSettings("display", "stackStyle", "spotlight");
    const b = mountBoard(gameView([spell("bolt", OPP, [{ kind: "card", id: "vivi" }])], [bolt]));
    const viviEl = b.q('.board [data-instance-id="vivi"]')!;
    expect(viviEl.hasAttribute(TARGET_ATTR)).toBe(true);
    b.destroy();
    expect(viviEl.hasAttribute(TARGET_ATTR)).toBe(false);
  });
});

describe("arrows while choosing targets (ADR 0119 §4)", () => {
  it("glows the source and draws one arrow per pick, from the source", () => {
    const b = mountBoard(gameView([]));
    expect(b.q(".targeting-arrows")).toBeNull();
    targeting.set(
      prompt({
        max: 3,
        picked: [
          { kind: "card", id: "birds" },
          { kind: "player", id: OPP },
        ],
      }),
    );
    flushSync();
    const shockEl = b.q('.board [data-instance-id="shock"]')!;
    expect(shockEl.hasAttribute("data-targeting-source")).toBe(true);
    const picks = b.qa('.targeting-arrows path[data-targeting-arrow="pick"]');
    expect(picks.map((p) => p.dataset.arrowTarget)).toEqual(["birds", OPP]);
    // Every arrow leaves the hand card's top edge.
    for (const p of picks) expect(p.getAttribute("d")).toMatch(/^M [\d.]+ 740 Q/);

    targeting.set(null);
    flushSync();
    expect(shockEl.hasAttribute("data-targeting-source")).toBe(false);
    expect(b.q(".targeting-arrows")).toBeNull();
  });

  it("starts from the dock's question line when the source is not on screen", () => {
    const dock = document.createElement("section");
    dock.setAttribute("aria-label", "actions");
    dock.dataset.testDock = "";
    dock.innerHTML = '<div class="dock-question">Select target for Flashback Bolt</div>';
    document.body.appendChild(dock);

    const b = mountBoard(gameView([]));
    targeting.set(
      prompt({
        card: card("gy-bolt", "Lightning Bolt", ME),
        picked: [{ kind: "card", id: "birds" }],
      }),
    );
    flushSync();
    expect(b.q("[data-targeting-source]")).toBeNull();
    const pick = b.q('.targeting-arrows path[data-targeting-arrow="pick"]')!;
    // The question line is at x 1180–1380, y 830–866; Birds is up and to the left.
    const [, x, y] = /^M ([\d.]+) ([\d.]+)/.exec(pick.getAttribute("d")!)!.map(Number);
    expect(x).toBeGreaterThanOrEqual(1180);
    expect(y).toBe(830);
  });

  it("draws no arrow to a graveyard pick", () => {
    const b = mountBoard(gameView([]));
    targeting.set(prompt({ picked: [{ kind: "card", id: "gy-card" }] }));
    flushSync();
    expect(b.qa('.targeting-arrows path[data-targeting-arrow="pick"]')).toHaveLength(0);
  });

  it("follows a mouse while a pick is open, and snaps green to a legal target", async () => {
    const b = mountBoard(gameView([]));
    targeting.set(prompt());
    flushSync();

    pointer("mouse", 200, 400);
    await nextFrame();
    let follow = b.q('.targeting-arrows path[data-targeting-arrow="follow"]');
    expect(follow).not.toBeNull();
    expect(follow!.classList.contains("snapped")).toBe(false);

    // Over Vivi, a legal target: the arrow snaps to it and turns green.
    const viviEl = b.q('.board [data-instance-id="vivi"]')!;
    document.elementFromPoint = () => viviEl;
    pointer("mouse", 550, 540);
    await nextFrame();
    follow = b.q('.targeting-arrows path[data-targeting-arrow="follow"]');
    expect(follow!.classList.contains("snapped")).toBe(true);
    expect(follow!.getAttribute("stroke")).toBe("#6fe3a4");

    // The clause fills: no pick is open, so the follow arrow goes.
    targeting.set(
      prompt({
        picked: [
          { kind: "card", id: "vivi" },
          { kind: "card", id: "birds" },
        ],
      }),
    );
    flushSync();
    expect(b.q('.targeting-arrows path[data-targeting-arrow="follow"]')).toBeNull();
    expect(b.qa('.targeting-arrows path[data-targeting-arrow="pick"]')).toHaveLength(2);
  });

  it("draws no follow arrow for a touch", async () => {
    const b = mountBoard(gameView([]));
    targeting.set(prompt());
    flushSync();
    pointer("touch", 200, 400);
    await nextFrame();
    expect(b.q('.targeting-arrows path[data-targeting-arrow="follow"]')).toBeNull();
    // The source still glows.
    expect(b.q('.board [data-instance-id="shock"]')?.hasAttribute("data-targeting-source")).toBe(
      true,
    );
  });
});
