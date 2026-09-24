// @vitest-environment jsdom
//
// handOrder.render.test.ts — #1524. handOrder.test.ts and
// dragCast.test.ts pin the decisions; this file pins what only the
// markup shows: that the hand renders in the saved order, that a drag
// sideways opens a gap, reorders and saves, that a drag up still casts,
// and that the sort menu is a real menu that reorders the hand.
//
// jsdom lays nothing out, so the hand's geometry is stubbed: the hand
// is the band y 500..700, and slot i spans x 100i..100i+100 (centre
// 100i+50). The cast line is 500 - CAST_ZONE_MARGIN_PX = 440.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

const preview = vi.hoisted(() => ({ calls: 0 }));

vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: () => {
      preview.calls++;
      return Promise.resolve({ ok: true, cost: "{R}", plan: [] });
    },
  };
});

vi.mock("./sounds", () => ({ play: () => {} }));

import Hand from "./components/board/Hand.svelte";
import type { CardView, GameView, LegalMoveView } from "./protocol";
import { handOrderKey } from "./handOrder";
import { render, cleanup, flushSync } from "./test/render.svelte";

const ME = "me";
const GAME = "g1";
const KEY = handOrderKey(GAME, ME);

const bolt: CardView = {
  instance_id: "bolt",
  name: "Lightning Bolt",
  owner: ME,
  controller: ME,
  type_line: "Instant",
  mana_cost: "{R}",
  colors: ["R"],
};
const bear: CardView = {
  instance_id: "bear",
  name: "Grizzly Bears",
  owner: ME,
  controller: ME,
  type_line: "Creature — Bear",
  mana_cost: "{1}{G}",
  colors: ["G"],
};
const forest: CardView = {
  instance_id: "forest",
  name: "Forest",
  owner: ME,
  controller: ME,
  type_line: "Basic Land — Forest",
};

function snap(moves: LegalMoveView[]): GameView {
  return {
    id: GAME,
    state: "active",
    seats: [
      { id: ME, name: "Me", hand: { kind: "hand", count: 0, cards: [] } },
      { id: "them", name: "Them", hand: { kind: "hand", count: 0, cards: [] } },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { seq: 4, number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    legal_moves: moves,
  } as unknown as GameView;
}

const castMove = (source: string): LegalMoveView => ({
  type: "cast_spell",
  player: ME,
  kind: "cast",
  label: "Cast",
  source,
});

const allCastable = snap([castMove("bolt"), castMove("bear")]);

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

let rectSpy: ReturnType<typeof vi.spyOn> | null = null;

beforeEach(() => {
  preview.calls = 0;
  localStorage.clear();
  rectSpy = vi.spyOn(Element.prototype, "getBoundingClientRect").mockImplementation(function (
    this: Element,
  ): DOMRect {
    const slot = this.closest(".hand-slot");
    if (slot) {
      const i = [...slot.parentElement!.querySelectorAll(":scope > .hand-slot")].indexOf(slot);
      return rect(i * 100, 500, 100, 200);
    }
    if (this.classList.contains("hand")) return rect(0, 500, 600, 200);
    return rect(0, 0, 0, 0);
  });
});

afterEach(() => {
  cleanup();
  rectSpy?.mockRestore();
  localStorage.clear();
});

function mount(cards: CardView[], isSelf = true, view: GameView = allCastable) {
  const played: string[] = [];
  const dragged: string[] = [];
  const r = render(
    Hand as never,
    {
      hand: { kind: "hand", owner: ME, count: cards.length, cards },
      isSelf,
      onPlayCard: (c: CardView) => played.push(c.instance_id),
      onDragCast: isSelf ? (c: CardView) => dragged.push(c.instance_id) : undefined,
      snap: view,
      viewerID: ME,
    } as never,
  );
  return { ...r, played, dragged };
}

function pointer(type: string, target: EventTarget, x: number, y: number): void {
  const Ctor = (
    typeof PointerEvent === "function" ? PointerEvent : MouseEvent
  ) as typeof MouseEvent;
  const ev = new Ctor(type, { bubbles: true, cancelable: true, clientX: x, clientY: y, button: 0 });
  Object.defineProperty(ev, "pointerId", { value: 1 });
  Object.defineProperty(ev, "isPrimary", { value: true });
  target.dispatchEvent(ev);
  flushSync();
}

const order = (c: HTMLElement): string[] =>
  [...c.querySelectorAll<HTMLElement>(".hand-slot .card")].map(
    (el) => el.getAttribute("aria-label") ?? "",
  );
const slotOf = (c: HTMLElement, name: string): HTMLElement =>
  c.querySelector<HTMLElement>(`.card[aria-label='${name}']`)!.closest<HTMLElement>(".hand-slot")!;
const ghost = () => document.body.querySelector<HTMLElement>(".drag-ghost");
const saved = (): string[] | null => {
  const raw = localStorage.getItem(KEY);
  return raw ? (JSON.parse(raw) as string[]) : null;
};

describe("rearranging your hand (#1524)", () => {
  it("renders in server order when nothing is saved, and saves nothing", () => {
    const { container } = mount([bolt, bear, forest]);
    expect(order(container)).toEqual(["Lightning Bolt", "Grizzly Bears", "Forest"]);
    expect(saved()).toBeNull();
  });

  it("dragging sideways opens a gap, reorders on release and saves", () => {
    const { container, dragged, played } = mount([bolt, bear, forest]);
    const slot = slotOf(container, "Forest"); // index 2, centre 250
    pointer("pointerdown", slot, 250, 600);
    pointer("pointermove", window, 200, 600);
    pointer("pointermove", window, 20, 610);

    // Reorder mode: a neutral ghost, no cast zone, the source dimmed,
    // and both other cards slid right of the gap at index 0.
    expect(ghost()?.classList.contains("reordering")).toBe(true);
    expect(ghost()?.classList.contains("castable")).toBe(false);
    expect(document.body.querySelector(".drag-cast-zone")).toBeNull();
    expect(slot.classList.contains("drag-source")).toBe(true);
    expect(container.querySelectorAll(".hand-slot.gap-right")).toHaveLength(2);
    expect(container.querySelector(".hand-slot.gap-left")).toBeNull();

    // Past the first card: the gap moves between them.
    pointer("pointermove", window, 120, 610);
    expect(slotOf(container, "Lightning Bolt").classList.contains("gap-left")).toBe(true);
    expect(slotOf(container, "Grizzly Bears").classList.contains("gap-right")).toBe(true);

    pointer("pointerup", window, 20, 610);
    slot.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    flushSync();

    expect(order(container)).toEqual(["Forest", "Lightning Bolt", "Grizzly Bears"]);
    expect(saved()).toEqual(["forest", "bolt", "bear"]);
    expect(dragged).toEqual([]);
    expect(played, "the drag swallowed its click").toEqual([]);
    expect(container.querySelector(".gap-left, .gap-right")).toBeNull();
    expect(preview.calls, "a reorder fetches no auto-tap preview").toBe(0);
  });

  it("dragging up past the line still casts, and does not reorder", async () => {
    const { container, dragged } = mount([bolt, bear, forest]);
    const slot = slotOf(container, "Grizzly Bears");
    pointer("pointerdown", slot, 150, 600);
    pointer("pointermove", window, 150, 580);
    expect(preview.calls).toBe(0);
    pointer("pointermove", window, 150, 400);
    await Promise.resolve();
    expect(preview.calls, "fetched once the card left the band").toBe(1);
    expect(document.body.querySelector(".drag-cast-zone")?.textContent).toContain(
      "Release to cast",
    );
    expect(ghost()?.classList.contains("castable")).toBe(true);
    pointer("pointerup", window, 150, 400);
    expect(dragged).toEqual(["bear"]);
    expect(order(container)).toEqual(["Lightning Bolt", "Grizzly Bears", "Forest"]);
    expect(saved()).toBeNull();
  });

  it("released between the band and the line, the card snaps back unmoved", () => {
    const { container, dragged } = mount([bolt, bear, forest]);
    const slot = slotOf(container, "Lightning Bolt");
    pointer("pointerdown", slot, 50, 600);
    pointer("pointermove", window, 400, 470);
    pointer("pointerup", window, 400, 470);
    expect(dragged).toEqual([]);
    expect(order(container)).toEqual(["Lightning Bolt", "Grizzly Bears", "Forest"]);
    expect(saved()).toBeNull();
  });

  it("a land can be dragged sideways into a new place", () => {
    const { container } = mount([forest, bolt, bear]);
    pointer("pointerdown", slotOf(container, "Forest"), 50, 600);
    pointer("pointermove", window, 290, 600);
    pointer("pointerup", window, 290, 600);
    expect(order(container)).toEqual(["Lightning Bolt", "Grizzly Bears", "Forest"]);
  });

  it("a saved order is applied on mount, with a new card appended on the right", () => {
    localStorage.setItem(KEY, JSON.stringify(["forest", "bolt"]));
    const { container } = mount([bolt, bear, forest]);
    expect(order(container)).toEqual(["Forest", "Lightning Bolt", "Grizzly Bears"]);
    // …and quietly recorded, so the new card keeps its place.
    expect(saved()).toEqual(["forest", "bolt", "bear"]);
  });

  it("a card that leaves the hand drops out of the saved order", async () => {
    // jsdom has no Web Animations; the leaving card's deal-out outro
    // needs one that finishes.
    const proto = Element.prototype as unknown as { animate?: unknown };
    const had = proto.animate;
    proto.animate = () => {
      let done: (() => void) | null = null;
      return {
        cancel() {},
        finish() {},
        currentTime: 0,
        get onfinish() {
          return done;
        },
        set onfinish(f: (() => void) | null) {
          done = f;
          setTimeout(() => f?.(), 0);
        },
      };
    };
    try {
      localStorage.setItem(KEY, JSON.stringify(["forest", "bear", "bolt"]));
      const r = mount([bolt, bear, forest]);
      r.setProps({ hand: { kind: "hand", owner: ME, count: 2, cards: [bolt, forest] } } as never);
      await new Promise((res) => setTimeout(res, 0));
      flushSync();
      expect(saved()).toEqual(["forest", "bolt"]);
      // The leaving card may still be playing its deal-out outro; the
      // cards that stayed keep the saved order.
      expect(order(r.container).filter((n) => n !== "Grizzly Bears")).toEqual([
        "Forest",
        "Lightning Bolt",
      ]);
    } finally {
      proto.animate = had;
    }
  });

  it("an order saved for another game is ignored", () => {
    localStorage.setItem(
      handOrderKey("other-game", ME),
      JSON.stringify(["forest", "bear", "bolt"]),
    );
    const { container } = mount([bolt, bear, forest]);
    expect(order(container)).toEqual(["Lightning Bolt", "Grizzly Bears", "Forest"]);
  });

  it("the sort menu is a real menu and sorts the hand once", () => {
    const { container } = mount([bolt, bear, forest]);
    const button = document.body.querySelector<HTMLButtonElement>(
      "button[aria-label='Sort hand']",
    )!;
    expect(button).not.toBeNull();
    expect(button.getAttribute("aria-haspopup")).toBe("menu");
    expect(button.getAttribute("aria-expanded")).toBe("false");
    button.click();
    flushSync();
    expect(button.getAttribute("aria-expanded")).toBe("true");
    const items = [...document.body.querySelectorAll<HTMLButtonElement>("[role='menuitem']")];
    expect(items.map((b) => b.textContent)).toEqual([
      "By mana value",
      "By type (lands first)",
      "By colour",
    ]);

    items[1].click();
    flushSync();
    expect(order(container)).toEqual(["Forest", "Grizzly Bears", "Lightning Bolt"]);
    expect(saved()).toEqual(["forest", "bear", "bolt"]);
    expect(document.body.querySelector("[role='menu']"), "the menu closes").toBeNull();

    // One-time: it can be adjusted by hand afterwards.
    pointer("pointerdown", slotOf(container, "Lightning Bolt"), 250, 600);
    pointer("pointermove", window, 20, 600);
    pointer("pointerup", window, 20, 600);
    expect(order(container)).toEqual(["Lightning Bolt", "Forest", "Grizzly Bears"]);
  });

  it("Escape closes the sort menu", () => {
    mount([bolt, bear]);
    const button = document.body.querySelector<HTMLButtonElement>(
      "button[aria-label='Sort hand']",
    )!;
    button.click();
    flushSync();
    const menu = document.body.querySelector<HTMLElement>("[role='menu']")!;
    menu.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(document.body.querySelector("[role='menu']")).toBeNull();
  });

  it("an opponent's hand has no sort menu and ignores a saved order", () => {
    localStorage.setItem(KEY, JSON.stringify(["forest", "bear", "bolt"]));
    const revealed = [bolt, bear, forest].map((c) => ({ ...c, known_by_you: true }));
    const { container } = mount(revealed, false);
    expect(document.body.querySelector("button[aria-label='Sort hand']")).toBeNull();
    expect(order(container)).toEqual(["Lightning Bolt", "Grizzly Bears", "Forest"]);
  });
});
