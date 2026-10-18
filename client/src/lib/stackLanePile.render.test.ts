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
    // ADR 0204 §4: the seats' avatars, at the board's right edge, away
    // from the pile's column, which keeps clear of any it overlaps.
    const left = this.hasAttribute("data-seat-identity") ? 1290 : 10;
    return {
      left,
      top,
      width: 100,
      height: 140,
      right: left + 100,
      bottom: top + 140,
      x: left,
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

  it("draws an ability as an ability card: its source's art crop and name, its kind, its text", () => {
    const p = mountPile(gameView([viviTrigger], { stackCards: [] }));
    const face = p.q(".top-card.ability .ability-card");
    expect(face?.dataset.abilityKind).toBe("triggered");
    expect(face?.querySelector(".kind-line")?.textContent).toBe("triggered ability");
    expect(face?.querySelector(".source-name")?.textContent).toBe("Vivi Ornitier");
    expect(face?.querySelector(".ability-text")?.textContent).toBe(
      "Vivi Ornitier — put a +1/+1 counter on Vivi",
    );
    const img = p.q(".top-card img");
    expect(img?.getAttribute("src")).toContain("sf-vivi");
    expect(img?.getAttribute("src")).toContain("art_crop");
    // The alt is the source's name and the ability's text (ADR 0119 §B).
    expect(img?.getAttribute("alt")).toBe(
      "Vivi Ornitier. Vivi Ornitier — put a +1/+1 counter on Vivi",
    );
    // A triggered ability has no cost badge (call 23).
    expect(face?.querySelector(".cost-badge")).toBeNull();
    // The card says the kind; the chips do not repeat it, nor the caption the title.
    expect(p.qa(".caption .chip").map((c) => c.textContent)).not.toContain("triggered");
    expect(p.q(".caption .top-title")).toBeNull();
  });

  it("badges a loyalty ability with its signed cost and an activated one with its symbols", () => {
    const sorin = card("sorin", "Sorin, Vampire Lord", OPP, {
      type_line: "Legendary Planeswalker — Sorin",
    });
    const minus = {
      id: "loy",
      kind: "activated",
      controller: OPP,
      owner: OPP,
      source_card_id: "sorin",
      label: "−2: Sorin deals 4 damage to any target. You gain 4 life.",
      ability_text: "−2: Sorin deals 4 damage to any target. You gain 4 life.",
    } as StackItemView;
    const view = gameView([minus], { stackCards: [] });
    (view.battlefield.cards as CardView[]).push(sorin);
    const p = mountPile(view);
    const face = p.q(".ability-card");
    expect(face?.dataset.abilityKind).toBe("loyalty");
    expect(face?.querySelector(".kind-line")?.textContent).toBe("loyalty ability");
    const badge = face?.querySelector<HTMLElement>(".cost-badge");
    expect(badge?.textContent?.trim()).toBe("\u22122");
    expect(badge?.dataset.costKind).toBe("loyalty");
    cleanup();

    const tap = {
      ...minus,
      id: "act",
      source_card_id: "vivi",
      label: "Vivi Ornitier — mana",
      ability_text: "{0}: Add {U} or {R}. Activate only during your turn and only once each turn.",
    } as StackItemView;
    const q = mountPile(gameView([tap], { stackCards: [] }));
    expect(q.q(".ability-card")?.dataset.abilityKind).toBe("activated");
    expect(q.q(".kind-line")?.textContent).toBe("activated ability");
    expect(q.q(".cost-badge .mana-cost")?.getAttribute("data-cost")).toBe("{0}");
  });

  it("draws a hidden source's ability on the card back's pattern with the model's name", () => {
    const faceDown = card("fd", "", OPP, { known_by_you: false, scryfall_id: undefined });
    const trig = {
      id: "fd-trig",
      kind: "triggered",
      controller: OPP,
      owner: OPP,
      source_card_id: "fd",
      label: "a face-down creature's trigger",
    } as StackItemView;
    const view = gameView([trig], { stackCards: [] });
    (view.battlefield.cards as CardView[]).push(faceDown);
    const p = mountPile(view);
    const img = p.q(".ability-card img");
    expect(img?.getAttribute("src")).toBe("/card-back.jpg");
    expect(p.q(".ability-card")?.classList.contains("hidden")).toBe(true);
    expect(p.q(".ability-card .source-name")?.textContent).toBe("face-down card");
    expect(p.q(".ability-card .ability-text")?.textContent).toBe("a face-down creature's trigger");
  });

  it("#3830: puts an ability's printed text on its card and its peeking line", () => {
    const printed =
      "Whenever Vivi Ornitier deals noncombat damage to a player, put a +1/+1 counter on it.";
    const withText = { ...viviTrigger, ability_text: printed };
    const top = mountPile(gameView([withText], { stackCards: [] }));
    expect(top.q(".top-card .ability-card .ability-text")?.textContent).toBe(printed);
    expect(top.q(".top-card")?.getAttribute("title")).toBe(`${printed} — hover to preview`);

    // Under a spell, the peeking line truncates it and its title has all of it.
    const under = mountPile(gameView([withText, spell("cs", OPP)]));
    const peek = under.q('.peek[data-stack-item-id="trig-vivi"]');
    expect(peek?.querySelector(".peek-title")?.textContent).toBe(printed);
    expect(peek?.getAttribute("title")).toBe(`${printed} — hover to preview`);
    // The Counter button keeps the label as its accessible name.
    expect(peek?.querySelector("button[aria-label^='Counter']")?.getAttribute("aria-label")).toBe(
      `Counter ${viviTrigger.label}`,
    );
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

describe("the cascade (ADR 0119 amendment §A)", () => {
  it("draws each lower item as a card: its face, its seat band, its title and Counter on the strip", () => {
    const p = mountPile(gameView(three));
    const bolt = p.q('.peek[data-stack-item-id="bolt"]')!;
    // A spell's face is its normal image, hidden from a screen reader:
    // the strip says it in text.
    const face = bolt.querySelector(".peek-face")!;
    expect(face.getAttribute("aria-hidden")).toBe("true");
    expect(face.querySelector("img")?.getAttribute("src")).toContain("sf-bolt");
    expect(face.querySelector("img")?.getAttribute("src")).not.toContain("art_crop");
    expect(bolt.querySelector(".seat-band")).not.toBeNull();
    expect(bolt.style.getPropertyValue("--seat-color")).not.toBe("");
    const strip = bolt.querySelector<HTMLElement>(".peek-line[data-stack-arrow-from]")!;
    expect(strip.querySelector(".peek-title")?.textContent).toBe("Lightning Bolt");
    expect(strip.querySelector(".peek-caster")?.textContent).toBe("You");
    expect(strip.querySelector(".counter-btn")?.getAttribute("aria-label")).toBe(
      "Counter Lightning Bolt",
    );
    // An ability's face is the ability card, hidden the same way.
    const trig = p.q('.peek[data-stack-item-id="trig-vivi"]')!;
    expect(trig.classList.contains("ability")).toBe(true);
    expect(trig.querySelector(".peek-face .ability-card")?.getAttribute("data-ability-kind")).toBe(
      "triggered",
    );
    expect(trig.querySelector(".peek-face")?.getAttribute("aria-hidden")).toBe("true");
    // Its strip shows its source's name, marked as an ability; the full
    // text stays the title (visually hidden), the tooltip and Counter's name.
    const source = trig.querySelector(".peek-source");
    expect(source?.textContent).toBe("Vivi Ornitier");
    expect(source?.getAttribute("aria-hidden")).toBe("true");
    expect(trig.querySelector(".peek-kind")).not.toBeNull();
    const title = trig.querySelector(".peek-title");
    expect(title?.textContent).toBe(viviTrigger.label);
    expect(title?.classList.contains("visually-hidden")).toBe(true);
    expect(trig.getAttribute("title")).toBe(`${viviTrigger.label} — hover to preview`);
    expect(trig.querySelector(".counter-btn")?.getAttribute("aria-label")).toBe(
      `Counter ${viviTrigger.label}`,
    );
    // A spell's strip is its title, as before.
    expect(bolt.querySelector(".peek-source")).toBeNull();
    expect(bolt.querySelector(".peek-title")?.classList.contains("visually-hidden")).toBe(false);
  });

  it("keeps the title on the strip of an ability whose source the viewer may not see", () => {
    const faceDown = card("fd", "", OPP, { known_by_you: false, scryfall_id: undefined });
    const trig = {
      id: "fd-trig",
      kind: "triggered",
      controller: OPP,
      owner: OPP,
      source_card_id: "fd",
      label: "a face-down creature's trigger",
    } as StackItemView;
    const view = gameView([trig, spell("cs", OPP)]);
    (view.battlefield.cards as CardView[]).push(faceDown);
    const p = mountPile(view);
    const peek = p.q('.peek[data-stack-item-id="fd-trig"]')!;
    expect(peek.querySelector(".peek-source")).toBeNull();
    const title = peek.querySelector(".peek-title");
    expect(title?.textContent).toBe("a face-down creature's trigger");
    expect(title?.classList.contains("visually-hidden")).toBe(false);
  });

  it("steps each card up, inward and smaller, dimmer the deeper it sits", () => {
    const items: StackItemView[] = [];
    const cards: CardView[] = [];
    for (let i = 0; i < 5; i++) {
      cards.push(card(`s${i}`, `Spell ${i}`, OPP, { type_line: "Instant" }));
      items.push(spell(`s${i}`, OPP));
    }
    const p = mountPile(gameView(items, { stackCards: cards }), {
      pile: { cardWidth: 250, shrunk: false },
    });
    const peeks = p.qa(".peek");
    // Deepest first in the DOM; each nearer one is drawn over it.
    expect(peeks.map((e) => e.dataset.depth)).toEqual(["4", "3", "2", "1"]);
    expect(peeks.map((e) => e.style.left)).toEqual(["60px", "45px", "30px", "15px"]);
    expect(peeks.map((e) => e.style.width)).toEqual(["210px", "220px", "230px", "240px"]);
    expect(peeks.map((e) => Number(e.style.zIndex))).toEqual([1, 2, 3, 4]);
    expect(peeks.map((e) => e.style.getPropertyValue("--dim"))).toEqual([
      "0.6",
      "0.6",
      "0.56",
      "0.28",
    ]);
    const tops = peeks.map((e) => parseFloat(e.style.top));
    expect(tops[0]).toBe(0);
    expect(tops[1] - tops[0]).toBeGreaterThan(0);
    // The pile is as wide as the cascade's overhang; the top card at its left.
    const pile = p.q(".pile")!;
    expect(pile.style.getPropertyValue("--pile-w")).toBe("270px");
    expect(pile.style.getPropertyValue("--top-left")).toBe("0px");
  });

  it("steps left when the host says the pile sits on the right half", () => {
    const p = mountPile(gameView(three), {
      pile: { cardWidth: 250, shrunk: false, inward: "left" },
    });
    const pile = p.q(".pile")!;
    expect(pile.classList.contains("inward-left")).toBe(true);
    expect(pile.style.getPropertyValue("--top-left")).toBe("10px");
    expect(p.qa(".peek").map((e) => e.style.left)).toEqual(["0px", "5px"]);
  });

  it("shows only as many lower cards as the host fits, the rest behind '+N more'", () => {
    const items: StackItemView[] = [];
    const cards: CardView[] = [];
    for (let i = 0; i < 5; i++) {
      cards.push(card(`s${i}`, `Spell ${i}`, OPP, { type_line: "Instant" }));
      items.push(spell(`s${i}`, OPP));
    }
    const p = mountPile(gameView(items, { stackCards: cards }), {
      pile: { cardWidth: 250, shrunk: false, lowerShown: 2 },
    });
    expect(p.qa(".peek-title").map((e) => e.textContent)).toEqual(["Spell 2", "Spell 3"]);
    expect(p.q(".more-chip")?.textContent?.trim()).toBe("+2 more");
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

describe("the top item's colour-identity edge (ADR 0119 amendment 2026-10-16 §C)", () => {
  it("edges the top card in its colours, and only the top card", () => {
    const blue = card("cs", "Counterspell", OPP, { type_line: "Instant", colors: ["U"] });
    const p = mountPile(gameView(three, { stackCards: [bolt, blue] }));
    const top = p.q(".top-card");
    expect(top?.classList.contains("edged")).toBe(true);
    expect(top?.style.getPropertyValue("--identity-edge")).toBe("var(--identity-u)");
    expect(p.qa(".top-card .identity-edge")).toHaveLength(1);
    expect(p.qa(".top-card .identity-glow")).toHaveLength(1);
    // Lower cards keep the seat band alone.
    expect(p.qa(".peek .identity-edge")).toHaveLength(0);
    // The face still holds the art, the seat band and the stack item ID.
    expect(top?.dataset.stackItemId).toBe("cs");
    expect(p.q(".top-card .top-face img")).not.toBeNull();
    expect(p.q(".top-card .top-face .seat-band")).not.toBeNull();
  });

  it("bands a multicolour item and gives a colourless one --identity-c", () => {
    const gold = card("cs", "Counterspell", OPP, { type_line: "Instant", colors: ["U", "W"] });
    const p = mountPile(gameView(three, { stackCards: [bolt, gold] }));
    const edge = p.q(".top-card")?.style.getPropertyValue("--identity-edge") ?? "";
    expect(edge).toMatch(/^conic-gradient\(from [\d.]+deg, var\(--identity-w\) 0deg/);
    expect(edge).toContain("var(--identity-u)");
    cleanup();
    const q = mountPile(gameView(three));
    expect(q.q(".top-card")?.style.getPropertyValue("--identity-edge")).toBe("var(--identity-c)");
  });

  it("edges an ability in its source's colours", () => {
    const red = { ...vivi, colors: ["R"] };
    const v = gameView([viviTrigger], { stackCards: [] });
    (v.battlefield as unknown as { cards: CardView[] }).cards = [red];
    const p = mountPile(v);
    expect(p.q(".top-card")?.style.getPropertyValue("--identity-edge")).toBe("var(--identity-r)");
  });

  it("keeps the accent ring, with no edge, for a card the viewer may not see", () => {
    const hidden = { ...counterspell, known_by_you: false, colors: ["U"] } as CardView;
    const p = mountPile(gameView(three, { stackCards: [bolt, hidden] }));
    expect(p.q(".top-card")?.classList.contains("edged")).toBe(false);
    expect(p.q(".top-card .identity-edge")).toBeNull();
  });
});
