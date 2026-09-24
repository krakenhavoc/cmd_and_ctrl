// @vitest-environment jsdom
//
// dragCast.render.test.ts — #1508. dragCast.test.ts pins the gesture's
// decisions; this file pins what only the markup can show: that a
// click on a hand card still casts through the click path, that a real
// drag puts a ghost on <body> with the gold / red class, dims the
// source card, opens the drop zone once past the line, hands the card
// to onDragCast on release and swallows the click that follows, and
// that the auto-tap preview drives the board highlight.
//
// jsdom lays nothing out, so every rect is zero: the hand's top edge is
// y = 0 and the table is anything above y = -CAST_ZONE_MARGIN_PX.

import { describe, it, expect, afterEach, vi, beforeEach } from "vitest";
import { get } from "svelte/store";

const preview = vi.hoisted(() => ({
  next: { ok: true, cost: "{R}", plan: ["mountain"] } as unknown,
  calls: 0,
}));

vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: () => {
      preview.calls++;
      return Promise.resolve(preview.next);
    },
  };
});

// jsdom has no audio; the hand's deal-in / play cues are not under test.
vi.mock("./sounds", () => ({ play: () => {} }));

import Hand from "./components/board/Hand.svelte";
import type { CardView, GameView, LegalMoveView } from "./protocol";
import { autoTapHighlight } from "./dragCast";
import { render, cleanup, flushSync } from "./test/render.svelte";

const ME = "me";

beforeEach(() => {
  preview.next = { ok: true, cost: "{R}", plan: ["mountain"] };
  preview.calls = 0;
});
afterEach(cleanup);

const bolt: CardView = {
  instance_id: "bolt",
  name: "Lightning Bolt",
  owner: ME,
  controller: ME,
  type_line: "Instant",
  mana_cost: "{R}",
};
const forest: CardView = {
  instance_id: "forest",
  name: "Forest",
  owner: ME,
  controller: ME,
  type_line: "Basic Land — Forest",
};

function snap(moves: LegalMoveView[], priority = 0): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      { id: ME, name: "Me", hand: { kind: "hand", count: 0, cards: [] } },
      { id: "them", name: "Them", hand: { kind: "hand", count: 0, cards: [] } },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: {
      seq: 4,
      number: 4,
      active_seat: 0,
      priority_holder: priority,
      phase: "main1",
      step: "main",
    },
    legal_moves: moves,
  } as unknown as GameView;
}

const move = (source: string, kind: "cast" | "land" = "cast"): LegalMoveView => ({
  type: "cast_spell",
  player: ME,
  kind,
  label: "Cast",
  source,
});

function mount(cards: CardView[], view: GameView, isSelf = true) {
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
  return { container: r.container, played, dragged };
}

function pointer(type: string, target: EventTarget, x: number, y: number): void {
  const Ctor = (
    typeof PointerEvent === "function" ? PointerEvent : MouseEvent
  ) as typeof MouseEvent;
  const ev = new Ctor(type, {
    bubbles: true,
    cancelable: true,
    clientX: x,
    clientY: y,
    button: 0,
  });
  Object.defineProperty(ev, "pointerId", { value: 1 });
  Object.defineProperty(ev, "isPrimary", { value: true });
  target.dispatchEvent(ev);
  flushSync();
}

const slotOf = (c: HTMLElement, name: string): HTMLElement =>
  c.querySelector<HTMLElement>(`.card[aria-label='${name}']`)!.closest<HTMLElement>(".hand-slot")!;
const ghost = () => document.body.querySelector<HTMLElement>(".drag-ghost");
const zone = () => document.body.querySelector<HTMLElement>(".drag-cast-zone");

// A drag onto the table: press, travel past the activation distance,
// then up above the line.
function dragToTable(slot: HTMLElement): void {
  pointer("pointerdown", slot, 10, 10);
  pointer("pointermove", window, 10, -20);
  pointer("pointermove", window, 10, -200);
}

describe("drag to cast (#1508)", () => {
  it("an ordinary click still casts through the click path", () => {
    const { container, played, dragged } = mount([bolt], snap([move("bolt")]));
    const slot = slotOf(container, "Lightning Bolt");
    expect(slot.classList.contains("draggable")).toBe(true);
    pointer("pointerdown", slot, 10, 10);
    pointer("pointerup", window, 12, 11);
    slot.querySelector<HTMLElement>(".card")!.click();
    flushSync();
    expect(played).toEqual(["bolt"]);
    expect(dragged).toEqual([]);
    expect(ghost()).toBeNull();
  });

  it("a drag shows a gold ghost, dims the source, and opens the zone past the line", () => {
    const { container } = mount([bolt], snap([move("bolt")]));
    const slot = slotOf(container, "Lightning Bolt");
    pointer("pointerdown", slot, 10, 10);
    pointer("pointermove", window, 10, -20);
    expect(ghost()?.classList.contains("castable")).toBe(true);
    expect(slot.classList.contains("drag-source")).toBe(true);
    expect(zone(), "short of the line: no zone yet").toBeNull();
    pointer("pointermove", window, 10, -200);
    expect(zone()?.textContent).toContain("Release to cast");
    pointer("pointerup", window, 10, -200);
  });

  it("releasing on the table hands the card to onDragCast and swallows the click", () => {
    const { container, played, dragged } = mount([bolt], snap([move("bolt")]));
    const slot = slotOf(container, "Lightning Bolt");
    dragToTable(slot);
    pointer("pointerup", window, 10, -200);
    // The browser's click for the same press, which must not also cast.
    slot.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    flushSync();
    expect(dragged).toEqual(["bolt"]);
    expect(played).toEqual([]);
    expect(ghost()).toBeNull();
    expect(zone()).toBeNull();
    expect(slot.classList.contains("drag-source")).toBe(false);
  });

  it("released short of the line, the card snaps back and nothing is cast", () => {
    const { container, dragged } = mount([bolt], snap([move("bolt")]));
    const slot = slotOf(container, "Lightning Bolt");
    pointer("pointerdown", slot, 10, 10);
    pointer("pointermove", window, 10, -30);
    pointer("pointerup", window, 10, -30);
    expect(dragged).toEqual([]);
    expect(document.body.querySelector("[role='status']")).toBeNull();
    const g = ghost();
    expect(g === null || g.classList.contains("snapping")).toBe(true);
  });

  it("an uncastable card goes red with the reason and snaps back showing it", () => {
    // Priority is with the other seat: the same Legality the hand greys with.
    const { container, dragged } = mount([bolt], snap([], 1));
    const slot = slotOf(container, "Lightning Bolt");
    dragToTable(slot);
    expect(ghost()?.classList.contains("blocked")).toBe(true);
    expect(ghost()?.textContent).toContain("Not your priority");
    pointer("pointerup", window, 10, -200);
    expect(dragged).toEqual([]);
    expect(document.body.querySelector("[role='status']")?.textContent).toContain(
      "Not your priority",
    );
    expect(preview.calls, "no preview for a card that can't be cast").toBe(0);
  });

  it("a land is not drag-played: red, 'Play lands by clicking', snaps back", () => {
    const { container, dragged } = mount([forest], snap([move("forest", "land")]));
    const slot = slotOf(container, "Forest");
    dragToTable(slot);
    expect(ghost()?.classList.contains("blocked")).toBe(true);
    pointer("pointerup", window, 10, -200);
    expect(dragged).toEqual([]);
    expect(document.body.querySelector("[role='status']")?.textContent).toContain(
      "Play lands by clicking",
    );
  });

  it("Escape cancels the drag", () => {
    const { container, dragged } = mount([bolt], snap([move("bolt")]));
    const slot = slotOf(container, "Lightning Bolt");
    dragToTable(slot);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    pointer("pointerup", window, 10, -200);
    expect(dragged).toEqual([]);
    expect(zone()).toBeNull();
  });

  it("the auto-tap preview highlights its sources while dragging, once per drag", async () => {
    const { container } = mount([bolt], snap([move("bolt")]));
    const slot = slotOf(container, "Lightning Bolt");
    dragToTable(slot);
    pointer("pointermove", window, 12, -210);
    await Promise.resolve();
    await Promise.resolve();
    expect(preview.calls).toBe(1);
    expect([...get(autoTapHighlight)]).toEqual(["mountain"]);
    pointer("pointerup", window, 10, -200);
    expect(get(autoTapHighlight).size).toBe(0);
  });

  it("a preview that can't pay turns the ghost red", async () => {
    preview.next = { ok: false, cost: "{R}", missing: ["{R}"] };
    const { container, dragged } = mount([bolt], snap([move("bolt")]));
    const slot = slotOf(container, "Lightning Bolt");
    dragToTable(slot);
    await Promise.resolve();
    await Promise.resolve();
    flushSync();
    expect(ghost()?.classList.contains("blocked")).toBe(true);
    expect(ghost()?.textContent).toContain("Can't pay {R}");
    pointer("pointerup", window, 10, -200);
    expect(dragged).toEqual([]);
  });

  it("an opponent's hand has no drag", () => {
    const { container } = mount([bolt], snap([move("bolt")]), false);
    expect(container.querySelector(".hand-slot.draggable")).toBeNull();
  });
});
