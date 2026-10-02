// @vitest-environment jsdom
//
// tutorialEmits.render.test.ts — ADR 0076 §2.5 (#1077). The three places a
// component tells the tutorial bus something the snapshot cannot: a card's
// ability popover opening, the viewer's hand being hovered, and a stacked
// pile of lands being hovered. Each must fire at that moment and no other,
// and a game with no subscriber must behave exactly as before.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import Card from "./components/board/Card.svelte";
import Hand from "./components/board/Hand.svelte";
import type { CardView, GameView } from "./protocol";
import { subscribe, type TutorialEvent } from "./tutorialBus";
import { render, cleanup, flushSync } from "./test/render.svelte";

let seen: TutorialEvent[] = [];
let off = () => {};

beforeEach(() => {
  seen = [];
  off = subscribe((n) => seen.push(n));
});

afterEach(() => {
  off();
  cleanup();
  localStorage.clear();
});

const land = (id: string): CardView => ({
  instance_id: id,
  name: "Forest",
  owner: "me",
  controller: "me",
  type_line: "Basic Land — Forest",
});

const spiritGuide = (): CardView =>
  ({
    instance_id: "c1",
    name: "Simian Spirit Guide",
    owner: "me",
    controller: "me",
    type_line: "Creature — Ape Spirit",
    zone_mana_abilities: [
      { index: 0, label: "Exile this card from your hand: Add {R}", produced: "{R}" },
    ],
  }) as unknown as CardView;

const enter = (el: Element): void => {
  const Ctor = (
    typeof PointerEvent === "function" ? PointerEvent : MouseEvent
  ) as typeof MouseEvent;
  // pointerenter does not bubble, which is what the handlers rely on.
  el.dispatchEvent(new Ctor("pointerenter", { bubbles: false }));
  flushSync();
};

const rightClick = (el: Element): void => {
  el.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
  flushSync();
};

describe("ability-menu-opened (Card.svelte)", () => {
  it("fires when right-click opens the popover, not when it closes it", () => {
    const view = render(
      Card as never,
      {
        card: spiritGuide(),
        onActivateManaAbility: () => {},
      } as never,
    );
    const root = view.container.querySelector<HTMLElement>(".card")!;
    rightClick(root);
    expect(view.container.querySelector('[role="menu"]')).toBeTruthy();
    expect(seen).toEqual(["ability-menu-opened"]);
    rightClick(root);
    expect(view.container.querySelector('[role="menu"]')).toBeNull();
    expect(seen).toEqual(["ability-menu-opened"]);
  });

  it("does not fire for a card with no menu to open", () => {
    const view = render(Card as never, { card: land("l1") } as never);
    rightClick(view.container.querySelector<HTMLElement>(".card")!);
    expect(seen).toEqual([]);
  });
});

describe("hand-hovered (Hand.svelte)", () => {
  const mountHand = (isSelf: boolean) =>
    render(
      Hand as never,
      {
        hand: { kind: "hand", owner: "me", count: 1, cards: [land("l1")] },
        isSelf,
        snap: { id: "g1", state: "active", seats: [], legal_moves: [] } as unknown as GameView,
        viewerID: "me",
      } as never,
    );

  it("fires when the pointer enters your own hand", () => {
    const { container } = mountHand(true);
    enter(container.querySelector('[aria-label="your hand"]')!);
    expect(seen).toEqual(["hand-hovered"]);
  });

  it("does not fire for an opponent's hand", () => {
    const { container } = mountHand(false);
    enter(container.querySelector('[aria-label="opponent hand"]')!);
    expect(seen).toEqual([]);
  });
});

describe("pile-hovered (BattlefieldRow.svelte)", () => {
  const mountRow = (cards: CardView[]) =>
    render(
      BattlefieldRow as never,
      { label: "lands", cards, viewerID: "me", strip: true } as never,
    );

  it("fires when the pointer enters a pile of several lands", () => {
    const { container } = mountRow([land("l1"), land("l2")]);
    enter(container.querySelector(".pile.multi")!);
    expect(seen).toEqual(["pile-hovered"]);
  });

  it("does not fire for a pile of one", () => {
    const { container } = mountRow([land("l1")]);
    enter(container.querySelector(".pile")!);
    expect(seen).toEqual([]);
  });
});

describe("with no subscriber", () => {
  it("every emit site runs without effect", () => {
    off();
    const row = render(
      BattlefieldRow as never,
      { label: "lands", cards: [land("l1"), land("l2")], viewerID: "me", strip: true } as never,
    );
    expect(() => enter(row.container.querySelector(".pile.multi")!)).not.toThrow();
    expect(seen).toEqual([]);
  });
});
