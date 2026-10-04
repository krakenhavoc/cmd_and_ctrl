// @vitest-environment jsdom
//
// ADR 0119 §1 — the pile, the default stack style. Placement and size
// are pinned as pure functions in stackPile.test.ts and the model in
// stackLane.test.ts; this file pins what only exists once the pile is
// rendered: the top card in front and the lower items peeking, the
// titles as text, the chips, Counter, "+N more", the pending group,
// the shrink while targeting, and on a real board the collapse tab,
// the phone fallback and the arrows.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import Board from "./components/board/Board.svelte";
import StackLanePile from "./components/board/StackLanePile.svelte";
import { settings, updateSettings } from "./settings";
import { targeting } from "./targeting";
import {
  buildStackLane,
  type StackLaneControls,
  type StackLaneItem,
  type StackPileSizing,
} from "./stackLane";
import { TARGET_ATTR } from "./stackArrows";
import type { ActionType, CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import { get } from "svelte/store";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

const ME = "me";
const OPP = "opp";

const realRect = Element.prototype.getBoundingClientRect;
const g = globalThis as Record<string, unknown>;
const realMatchMedia = g.matchMedia;

beforeEach(() => {
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  // jsdom does no layout. Give the board a desktop size (a short board
  // draws the compact card instead), every other element one small
  // box, and board cards a place below it, so a board target is not
  // "under" the pile and gets an arrow.
  Element.prototype.getBoundingClientRect = function (this: Element) {
    if (this.classList.contains("board")) {
      return {
        left: 0,
        top: 0,
        width: 1400,
        height: 900,
        right: 1400,
        bottom: 900,
        x: 0,
        y: 0,
        toJSON: () => ({}),
      } as DOMRect;
    }
    const top = this.hasAttribute("data-instance-id") ? 700 : 10;
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
  g.matchMedia = realMatchMedia;
  targeting.set(null);
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
    scryfall_id: `sf-${id}`,
    ...extra,
  }) as CardView;

const vivi = card("vivi", "Vivi Ornitier", ME, {
  type_line: "Legendary Creature — Wizard",
  power: 0,
  toughness: 3,
});
const bolt = card("bolt", "Lightning Bolt", ME, { type_line: "Instant" });
const counterspell = card("cs", "Counterspell", OPP, { type_line: "Instant" });
const growth = card("growth", "Giant Growth", OPP, { type_line: "Instant", unimplemented: true });

const spell = (id: string, controller: string, targets: StackItemView["targets"] = []) =>
  ({
    id,
    kind: "spell",
    controller,
    owner: controller,
    source_card_id: id,
    targets,
  }) as StackItemView;

const viviTrigger: StackItemView = {
  id: "trig-vivi",
  kind: "triggered",
  controller: ME,
  owner: ME,
  source_card_id: "vivi",
  label: "Vivi Ornitier — put a +1/+1 counter on Vivi",
};

// Bottom first, as the wire sends it: the pile reads Counterspell (top),
// Lightning Bolt, then Vivi's trigger.
const three: StackItemView[] = [
  viviTrigger,
  spell("bolt", ME, [{ kind: "card", id: "vivi" }]),
  spell("cs", OPP, [{ kind: "card", id: "bolt" }]),
];

const pendingTrigger: StackItemView = {
  id: "pend-1",
  kind: "triggered",
  controller: OPP,
  owner: OPP,
  source_card_id: "vivi",
  label: "Vivi Ornitier — waiting trigger",
};

const gameView = (
  items: StackItemView[],
  opts: { pending?: StackItemView[]; stackCards?: CardView[]; holder?: number } = {},
): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Bot 2", 1)],
    battlefield: zone("battlefield", undefined, [vivi]),
    stack: zone("stack", undefined, opts.stackCards ?? [bolt, counterspell]),
    exile: zone("exile", undefined),
    stack_items: items,
    pending_triggers: opts.pending ?? [],
    pending_choices: [],
    turn: {
      number: 3,
      active_seat: 0,
      priority_holder: opts.holder ?? 0,
      phase: "main1",
      step: "main",
    },
    mulligans_open: false,
  }) as unknown as GameView;

function model(view: GameView, viewerHolds = true) {
  return buildStackLane({
    stack: view.stack,
    stackItems: view.stack_items,
    pendingTriggers: view.pending_triggers,
    seats: view.seats,
    battlefield: view.battlefield,
    exile: view.exile,
    viewerID: ME,
    priorityHolder: viewerHolds ? ME : OPP,
    splitSecondActive: false,
    oracleTextFor: (c) =>
      c.name === "Counterspell"
        ? "Counter target spell."
        : c.name === "Vivi Ornitier"
          ? "Flying"
          : null,
  });
}

function fakeControls(targetable: (it: StackLaneItem) => boolean = () => false) {
  const calls = { counter: [] as string[], target: [] as string[], enter: [] as string[] };
  const controls: StackLaneControls = {
    counter: (it) => calls.counter.push(it.id),
    targetable,
    target: (it) => calls.target.push(it.id),
    hoverEnter: (it) => calls.enter.push(it.id),
    hoverLeave: () => {},
  };
  return { controls, calls };
}

function mountPile(
  view: GameView,
  opts: {
    viewerHolds?: boolean;
    pile?: StackPileSizing;
    targetable?: (it: StackLaneItem) => boolean;
  } = {},
) {
  const { controls, calls } = fakeControls(opts.targetable);
  const r = render(
    StackLanePile as never,
    {
      model: model(view, opts.viewerHolds ?? true),
      controls,
      styleName: "pile",
      pile: opts.pile ?? { cardWidth: 270, shrunk: false },
    } as never,
  );
  const qa = (sel: string) => Array.from(r.container.querySelectorAll<HTMLElement>(sel));
  const q = (sel: string) => r.container.querySelector<HTMLElement>(sel);
  return { ...r, calls, qa, q };
}

describe("the pile (ADR 0119 §1)", () => {
  it("puts the top item in front and the lower ones peeking above it, deepest first", () => {
    const p = mountPile(gameView(three));
    expect(p.q(".top-card")?.dataset.stackItemId).toBe("cs");
    expect(p.qa(".peek").map((e) => e.dataset.stackItemId)).toEqual(["trig-vivi", "bolt"]);
    // The titles are text, verbatim — e2e specs read them.
    expect(p.q(".top-title")?.textContent).toBe("Counterspell");
    expect(p.qa(".peek-title").map((e) => e.textContent)).toEqual([
      "Vivi Ornitier — put a +1/+1 counter on Vivi",
      "Lightning Bolt",
    ]);
  });

  it("draws the top card's normal image with its name and oracle text as the alt", () => {
    const p = mountPile(gameView(three));
    const img = p.q(".top-card img") as HTMLImageElement;
    expect(img.getAttribute("src")).toContain("sf-cs");
    expect(img.getAttribute("src")).not.toContain("art_crop");
    expect(img.getAttribute("alt")).toBe("Counterspell. Counter target spell.");
    expect(p.q(".top-card")?.style.getPropertyValue("--seat-color")).not.toBe("");
  });

  it("captions the top card with its caster above the title and the target line", () => {
    const p = mountPile(gameView(three));
    expect(p.q(".caption .caster")?.textContent).toBe("Bot 2");
    expect(p.q(".caption .aim")?.textContent).toBe("→ your Lightning Bolt");
    // The peeking line another item aims at is ringed.
    expect(p.q('.peek[data-stack-item-id="bolt"]')?.classList.contains("targeted")).toBe(true);
  });

  it("draws an ability as its source's card with a band carrying its label and kind", () => {
    const p = mountPile(gameView([viviTrigger], { stackCards: [] }));
    const band = p.q(".top-card .ability-band");
    expect(band?.querySelector(".kind-tag")?.textContent).toBe("triggered ability");
    expect(band?.querySelector(".title")?.textContent).toBe(
      "Vivi Ornitier — put a +1/+1 counter on Vivi",
    );
    expect(p.q(".top-card img")?.getAttribute("src")).toContain("sf-vivi");
    expect(p.q(".top-card img")?.getAttribute("alt")).toBe("Vivi Ornitier. Flying");
    // The band says the kind; the chips do not repeat it.
    expect(p.qa(".caption .chip").map((c) => c.textContent)).not.toContain("triggered");
  });

  it("draws a card back for a spell the viewer may not see", () => {
    const hidden = card("cs", "", OPP, { known_by_you: false, scryfall_id: undefined });
    const p = mountPile(gameView([spell("cs", OPP)], { stackCards: [hidden] }));
    const img = p.q(".top-card img");
    expect(img?.getAttribute("src")).toBe("/card-back.jpg");
    expect(img?.getAttribute("alt")).toBe("card back");
  });

  it("shows every chip the model gives, manual included", () => {
    const p = mountPile(gameView([spell("growth", OPP)], { stackCards: [growth] }));
    const chips = p.qa(".caption .chip");
    expect(chips.map((c) => c.textContent)).toContain("manual");
    expect(chips.find((c) => c.textContent === "manual")?.classList.contains("manual")).toBe(true);
  });

  it("puts Counter on the top card and on every peeking line, disabled without priority", () => {
    const p = mountPile(gameView(three));
    const btns = p.qa(".counter-btn");
    expect(btns.map((b) => b.getAttribute("aria-label"))).toEqual([
      "Counter Vivi Ornitier — put a +1/+1 counter on Vivi",
      "Counter Lightning Bolt",
      "Counter Counterspell",
    ]);
    expect(btns.every((b) => b.textContent?.trim() === "Counter")).toBe(true);
    click(btns[2]);
    expect(p.calls.counter).toEqual(["cs"]);
    cleanup();
    const q = mountPile(gameView(three), { viewerHolds: false });
    expect(q.qa(".counter-btn").every((b) => (b as HTMLButtonElement).disabled)).toBe(true);
  });

  it("hides a fifth lower item and deeper behind '+N more', which lists the whole stack", () => {
    const items: StackItemView[] = [];
    const cards: CardView[] = [];
    for (let i = 0; i < 7; i++) {
      cards.push(card(`s${i}`, `Spell ${i}`, OPP, { type_line: "Instant" }));
      items.push(spell(`s${i}`, OPP));
    }
    const p = mountPile(gameView(items, { stackCards: cards }));
    expect(p.qa(".peek")).toHaveLength(4);
    const chip = p.q('button[aria-label="show 2 more on the stack"]')!;
    expect(chip.textContent?.trim()).toBe("+2 more");
    expect(p.q(".all-items")).toBeNull();
    click(chip);
    const rows = p.qa(".all-items .row-title").map((r) => r.textContent);
    expect(rows).toEqual([
      "Spell 6",
      "Spell 5",
      "Spell 4",
      "Spell 3",
      "Spell 2",
      "Spell 1",
      "Spell 0",
    ]);
    expect(chip.getAttribute("aria-expanded")).toBe("true");
    // A second click closes it…
    click(chip);
    expect(p.q(".all-items")).toBeNull();
    // …and so does Escape.
    click(chip);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    return Promise.resolve().then(() => expect(p.q(".all-items")).toBeNull());
  });

  it("shows no '+N more' at five items", () => {
    const items: StackItemView[] = [];
    const cards: CardView[] = [];
    for (let i = 0; i < 5; i++) {
      cards.push(card(`s${i}`, `Spell ${i}`, OPP));
      items.push(spell(`s${i}`, OPP));
    }
    const p = mountPile(gameView(items, { stackCards: cards }));
    expect(p.qa(".peek")).toHaveLength(4);
    expect(p.q(".more-chip")).toBeNull();
  });

  it("lists pending triggers under the pile in a group", () => {
    const p = mountPile(gameView(three, { pending: [pendingTrigger] }));
    const group = p.q('[role="group"][aria-label="waiting to go on the stack"]');
    expect(group).not.toBeNull();
    expect(group?.textContent).toContain("Vivi Ornitier — waiting trigger");
    // Pending triggers are not on the stack: no Counter for them.
    expect(group?.querySelector(".counter-btn")).toBeNull();
  });

  it("shrinks to its top card at 160px while the viewer chooses a target", () => {
    const p = mountPile(gameView(three, { pending: [pendingTrigger] }), {
      pile: { cardWidth: 270, shrunk: true },
    });
    expect(p.q(".pile")?.style.getPropertyValue("--card-w")).toBe("160px");
    expect(p.qa(".peek")).toHaveLength(0);
    expect(p.q(".top-card")?.dataset.stackItemId).toBe("cs");
    expect(p.q('[aria-label="waiting to go on the stack"]')).toBeNull();
  });

  it("offers a Target button on an item a prompt may point at", () => {
    const p = mountPile(gameView(three), { targetable: (it) => it.id === "bolt" });
    const btn = p.q('.peek .target-btn[aria-label="Target Lightning Bolt"]')!;
    click(btn);
    expect(p.calls.target).toEqual(["bolt"]);
    expect(p.qa(".target-btn")).toHaveLength(1);
  });
});

describe("the pile on a real board", () => {
  function mountBoard(view: GameView) {
    const sent: { type: ActionType; params?: unknown }[] = [];
    const r = render(
      Board as never,
      {
        view,
        viewerID: ME,
        isAdmin: false,
        sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
      } as never,
    );
    const q = (sel: string) => r.container.querySelector<HTMLElement>(sel);
    return { ...r, sent, q };
  }

  it("is what the default setting draws, labelled as the stack, with no docked card", () => {
    updateSettings("display", "stackStyle", "pile");
    const b = mountBoard(gameView(three));
    expect(b.q('.lane-host[data-stack-style="pile"] [data-stack-body="pile"]')).not.toBeNull();
    expect(b.q(".stack-lane")?.getAttribute("aria-label")).toBe("stack: 3 on the stack");
    expect(b.q(".strip .overlay")).toBeNull();
    expect(b.container.querySelectorAll('[role="region"][aria-label="attention"]')).toHaveLength(1);
  });

  it("draws the compact card on a phone, leaving the setting alone", () => {
    g.matchMedia = (q: string) => ({
      matches: q.includes("599px"),
      media: q,
      addEventListener() {},
      removeEventListener() {},
    });
    updateSettings("display", "stackStyle", "pile");
    const b = mountBoard(gameView(three));
    expect(b.q(".stack-lane")).toBeNull();
    expect(b.q(".strip .overlay")).not.toBeNull();
    expect(get(settings).display.stackStyle).toBe("pile");
  });

  it("folds to a tab and back, and a new item on an empty stack starts expanded", () => {
    updateSettings("display", "stackStyle", "pile");
    const b = mountBoard(gameView(three));
    click(b.q('button[aria-label="hide the stack"]')!);
    const tab = b.q('button[aria-label="show the stack"]')!;
    expect(tab.textContent?.trim()).toBe("Stack · 3");
    expect(b.q('[data-stack-body="pile"]')).toBeNull();
    // Still labelled as the stack while folded.
    expect(b.q(".stack-lane")?.getAttribute("aria-label")).toBe("stack: 3 on the stack");
    click(tab);
    expect(b.q('[data-stack-body="pile"]')).not.toBeNull();

    click(b.q('button[aria-label="hide the stack"]')!);
    b.setProps({ view: gameView([]) } as never);
    b.setProps({ view: gameView([spell("bolt", ME)]) } as never);
    expect(b.q('[data-stack-body="pile"]')).not.toBeNull();
  });

  it("shrinks while the viewer chooses a board target, not while a stack item is one", () => {
    updateSettings("display", "stackStyle", "pile");
    const b = mountBoard(gameView(three));
    expect(b.container.querySelectorAll(".peek")).toHaveLength(2);
    targeting.set({
      card: bolt,
      mode: "any",
      legal: { players: new Set(), cards: new Set(["vivi"]) },
      min: 1,
      max: 1,
      picked: [],
      steps: [],
      step: 0,
      done: [],
    } as never);
    return Promise.resolve().then(() => {
      expect(b.container.querySelectorAll(".peek")).toHaveLength(0);
      targeting.set({
        card: bolt,
        mode: "any",
        legal: { players: new Set(), cards: new Set(["bolt"]) },
        min: 1,
        max: 1,
        picked: [],
        steps: [],
        step: 0,
        done: [],
      } as never);
      return Promise.resolve().then(() => {
        expect(b.container.querySelectorAll(".peek")).toHaveLength(2);
      });
    });
  });

  it("draws an arrow to a permanent it targets and rings it, and CombatArrows stands down", () => {
    updateSettings("display", "stackStyle", "pile");
    const b = mountBoard(gameView(three));
    const viviCard = b.q('.board [data-instance-id="vivi"]');
    expect(viviCard).not.toBeNull();
    const arrow = b.q('.lane-host .board-arrows path[data-arrow-target="vivi"]');
    expect(arrow).not.toBeNull();
    expect(arrow?.getAttribute("marker-end")).toMatch(/^url\(#stack-pile-/);
    expect(viviCard?.hasAttribute(TARGET_ATTR)).toBe(true);
    b.setProps({ view: gameView([]) } as never);
    expect(b.q('.board [data-instance-id="vivi"]')?.hasAttribute(TARGET_ATTR)).toBe(false);
  });

  it("Counter on the pile sends the same verb the docked card does", () => {
    updateSettings("display", "stackStyle", "pile");
    const b = mountBoard(gameView(three));
    click(b.q('.stack-lane .counter-btn[aria-label="Counter Counterspell"]')!);
    expect(b.sent).toEqual([{ type: "counter_spell", params: { instance_id: "cs" } }]);
  });
});
