// @vitest-environment jsdom
//
// exileDragCast.render.test.ts — #1622. Drag-to-cast from the exile
// strip: a card the viewer holds a permission over can be dragged onto
// the table and is handed to the same cast entry point a click uses
// (zone "exile", the grant's face); a card they hold no permission over
// is not in the strip, so nothing on it is draggable. Click-to-cast
// keeps working. jsdom lays nothing out, so the strip row's top edge is
// y = 0 and the table is anything above y = -CAST_ZONE_MARGIN_PX.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: () => Promise.resolve({ ok: true, cost: "{R}", plan: [] }),
  };
});

import ExileStrip from "./components/board/ExileStrip.svelte";
import type { CardView, GameView, LegalMoveView } from "./protocol";
import type { CastSourceZone } from "./targeting";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const THEM = "them";

function snap(cards: CardView[], moves: LegalMoveView[]): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      { id: ME, name: "Me", hand: { kind: "hand", count: 0, cards: [] } },
      { id: THEM, name: "Them", hand: { kind: "hand", count: 0, cards: [] } },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: cards.length, cards },
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

interface Handed {
  card: CardView;
  zone: CastSourceZone;
  face?: number;
}

function mount(view: GameView) {
  const clicked: Handed[] = [];
  const dragged: Handed[] = [];
  const r = render(
    ExileStrip as never,
    {
      view,
      viewerID: ME,
      onCastCard: (card: CardView, zone: CastSourceZone, face?: number) =>
        clicked.push({ card, zone, face }),
      onDragCast: (card: CardView, zone: CastSourceZone, face?: number) =>
        dragged.push({ card, zone, face }),
    } as never,
  );
  return { container: r.container, clicked, dragged };
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

const slotOf = (c: HTMLElement, name: string): HTMLElement =>
  c.querySelector<HTMLElement>(`.card[aria-label='${name}']`)!.closest<HTMLElement>(".strip-slot")!;
const ghost = () => document.body.querySelector<HTMLElement>(".drag-ghost");

function dragToTable(slot: HTMLElement): void {
  pointer("pointerdown", slot, 10, 10);
  pointer("pointermove", window, 10, -20);
  pointer("pointermove", window, 10, -200);
}

// Vivi, airbent: another seat's card, granted to the viewer, naming a face.
const airbent: CardView = {
  instance_id: "vivi",
  name: "Vivi",
  owner: THEM,
  controller: THEM,
  type_line: "Creature — Wizard",
  mana_cost: "{1}{U}",
  exile_play: { player: ME, faces: [0], cast_only: true },
  castable_here: true,
  cast_prices: [{ cost: "{2}" }],
};

// Exiled, but the grant is someone else's and nothing is castable for us.
const notMine: CardView = {
  instance_id: "theirs",
  name: "Their Bolt",
  owner: THEM,
  controller: THEM,
  type_line: "Instant",
  mana_cost: "{R}",
  exile_play: { player: THEM, cast_only: true },
};

describe("ExileStrip drag to cast — #1622", () => {
  it("dragging a card you hold a permission for hands it to the cast entry point", () => {
    const { container, clicked, dragged } = mount(snap([airbent], [castMove("vivi")]));
    const slot = slotOf(container, "Vivi");
    expect(slot.classList.contains("draggable")).toBe(true);
    dragToTable(slot);
    expect(ghost()?.classList.contains("castable")).toBe(true);
    expect(document.body.querySelector(".drag-cast-zone")?.textContent).toContain(
      "Release to cast",
    );
    pointer("pointerup", window, 10, -200);
    // The browser's click for the same press must not also cast.
    slot.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    flushSync();
    expect(dragged).toHaveLength(1);
    expect(dragged[0].zone).toBe("exile");
    expect(dragged[0].face).toBe(0);
    expect(dragged[0].card.instance_id).toBe("vivi");
    expect(clicked).toEqual([]);
    expect(ghost()).toBeNull();
  });

  it("a card you hold no permission for is not in the strip, so it is not draggable", () => {
    const { container, dragged } = mount(snap([notMine], [castMove("theirs")]));
    expect(container.querySelector(".strip-slot")).toBeNull();
    expect(container.querySelector(".draggable")).toBeNull();
    pointer("pointerdown", container, 10, 10);
    pointer("pointermove", window, 10, -200);
    pointer("pointerup", window, 10, -200);
    expect(dragged).toEqual([]);
  });

  it("releasing a card that cannot be cast snaps back with the reason and casts nothing", () => {
    const { container, dragged } = mount(snap([airbent], [castMove("something-else")]));
    const slot = slotOf(container, "Vivi");
    dragToTable(slot);
    expect(ghost()?.classList.contains("blocked")).toBe(true);
    pointer("pointerup", window, 10, -200);
    expect(dragged).toEqual([]);
    expect(document.body.querySelector("[role='status']")).not.toBeNull();
  });

  it("released short of the line, nothing is cast", () => {
    const { container, dragged } = mount(snap([airbent], [castMove("vivi")]));
    const slot = slotOf(container, "Vivi");
    pointer("pointerdown", slot, 10, 10);
    pointer("pointermove", window, 10, -30);
    pointer("pointerup", window, 10, -30);
    expect(dragged).toEqual([]);
  });

  it("a click still casts through onCastCard and never starts a drag", () => {
    const { container, clicked, dragged } = mount(snap([airbent], [castMove("vivi")]));
    const slot = slotOf(container, "Vivi");
    pointer("pointerdown", slot, 10, 10);
    pointer("pointerup", window, 12, 11);
    click(container.querySelector<HTMLElement>(".card[aria-label='Vivi']")!);
    expect(clicked).toHaveLength(1);
    expect(clicked[0].zone).toBe("exile");
    expect(clicked[0].face).toBe(0);
    expect(dragged).toEqual([]);
  });

  it("Escape cancels a drag in flight", () => {
    const { container, dragged } = mount(snap([airbent], [castMove("vivi")]));
    const slot = slotOf(container, "Vivi");
    dragToTable(slot);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    pointer("pointerup", window, 10, -200);
    expect(dragged).toEqual([]);
  });
});
