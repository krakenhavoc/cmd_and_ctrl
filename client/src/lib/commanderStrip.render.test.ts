// @vitest-environment jsdom
//
// commanderStrip.render.test.ts — #2202. The viewer's commander sits in
// the castable-from-other-zones strip beside the hand. A click hands it
// to the Board's cast chain with the command zone as its zone — the
// same call the command zone panel's own click makes — and a drag past
// the line hands it over with the drag flag (strict + auto-tap on the
// wire), the hand's gesture. The narrow-panel chip counts it.

import { describe, it, expect, afterEach, vi } from "vitest";

const previewCalls: unknown[][] = [];
vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: (...args: unknown[]) => {
      previewCalls.push(args);
      return Promise.resolve({ ok: true, cost: "{4}{G}", plan: [] });
    },
  };
});

import ExileStrip from "./components/board/ExileStrip.svelte";
import CommandZone from "./components/board/CommandZone.svelte";
import type { CardView, GameView, LegalMoveView } from "./protocol";
import { applyCastChoices, castChoicesBase, type CastSourceZone } from "./targeting";
import { legalActionsOf, type LegalActions } from "./legalActions";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  previewCalls.length = 0;
});

const ME = "me";
const THEM = "them";

const kenrith: CardView = {
  instance_id: "kenrith",
  name: "Kenrith",
  owner: ME,
  controller: ME,
  is_commander: true,
  type_line: "Legendary Creature — Human Noble",
  mana_cost: "{4}{W}",
  cast_prices: [{ cost: "{6}{W}" }],
} as CardView;

const theirCommander: CardView = {
  instance_id: "theirs",
  name: "Their Commander",
  owner: THEM,
  controller: THEM,
  is_commander: true,
  type_line: "Legendary Creature — Elf",
  mana_cost: "{G}",
} as CardView;

const exiled: CardView = {
  instance_id: "bolt",
  name: "Impulse Bolt",
  owner: THEM,
  controller: THEM,
  type_line: "Instant",
  mana_cost: "{R}",
  exile_play: { player: ME },
  castable_here: true,
  cast_prices: [{ cost: "{R}", printed: true }],
} as CardView;

function snap(opts: {
  mine?: CardView[];
  exile?: CardView[];
  moves?: LegalMoveView[];
  casts?: Record<string, number>;
}): GameView {
  const mine = opts.mine ?? [];
  return {
    id: "g",
    state: "active",
    seats: [
      {
        id: ME,
        name: "Me",
        hand: { kind: "hand", count: 0, cards: [] },
        command: { kind: "command", count: mine.length, cards: mine },
        commander_casts: opts.casts,
      },
      {
        id: THEM,
        name: "Them",
        hand: { kind: "hand", count: 0, cards: [] },
        command: { kind: "command", count: 1, cards: [theirCommander] },
      },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: (opts.exile ?? []).length, cards: opts.exile ?? [] },
    turn: { seq: 4, number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    legal_moves: opts.moves,
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

function mountStrip(view: GameView, legal?: LegalActions) {
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
      ...(legal ? { legal } : {}),
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

const cardEl = (c: HTMLElement, name: string) =>
  c.querySelector<HTMLElement>(`.card[aria-label='${name}']`);
const slotOf = (c: HTMLElement, name: string): HTMLElement =>
  cardEl(c, name)!.closest<HTMLElement>(".strip-slot")!;

describe("the commander in the strip — #2202", () => {
  it("sits in the strip, before the exile cards; an opponent's commander does not", () => {
    const { container } = mountStrip(
      snap({ mine: [kenrith], exile: [exiled], moves: [castMove("kenrith")] }),
    );
    const names = Array.from(container.querySelectorAll(".strip-slot .card")).map((c) =>
      c.getAttribute("aria-label"),
    );
    expect(names).toEqual(["Kenrith", "Impulse Bolt"]);
    expect(cardEl(container, "Their Commander")).toBeNull();
    expect(container.querySelector("[aria-label='castable from other zones']")).not.toBeNull();
  });

  it("a click starts the same cast as the command zone panel's click", () => {
    const view = snap({ mine: [kenrith], moves: [castMove("kenrith")] });
    const strip = mountStrip(view);
    click(cardEl(strip.container, "Kenrith")!);
    expect(strip.clicked).toHaveLength(1);

    const panel: { card: CardView; zone: CastSourceZone }[] = [];
    const sent: unknown[] = [];
    const cz = render(
      CommandZone as never,
      {
        seat: { id: ME, name: "Me" },
        zone: view.seats[0].command,
        isSelf: true,
        sendAction: (...a: unknown[]) => sent.push(a),
        view,
        viewerID: ME,
        onCastCard: (card: CardView, zone: CastSourceZone) => panel.push({ card, zone }),
      } as never,
    );
    click(cz.container.querySelector<HTMLElement>("[aria-label^='cast commander']")!);
    expect(sent).toEqual([]);
    expect(panel).toHaveLength(1);

    // Same card, same zone, no face: the chain gets one call either way.
    expect(strip.clicked[0].card.instance_id).toBe(panel[0].card.instance_id);
    expect(strip.clicked[0].zone).toBe("command");
    expect(panel[0].zone).toBe("command");
    expect(strip.clicked[0].face).toBeUndefined();

    // …and for a plain commander the chain's payload is the bare one the
    // command zone used to send itself.
    const params: Record<string, unknown> = { instance_id: "kenrith" };
    applyCastChoices(params, castChoicesBase("command"));
    expect(params).toEqual({ instance_id: "kenrith", from_zone: "command" });
  });

  it("greys a commander the move list cannot cast, and a click does nothing", () => {
    const { container, clicked } = mountStrip(
      snap({ mine: [kenrith], moves: [castMove("something-else")] }),
    );
    expect(slotOf(container, "Kenrith").classList.contains("blocked")).toBe(true);
    click(cardEl(container, "Kenrith")!);
    expect(clicked).toEqual([]);
  });

  it("dragged past the line, it casts from the command zone with strict + auto-tap", () => {
    const { container, clicked, dragged } = mountStrip(
      snap({ mine: [kenrith], moves: [castMove("kenrith")] }),
    );
    const slot = slotOf(container, "Kenrith");
    expect(slot.classList.contains("draggable")).toBe(true);
    pointer("pointerdown", slot, 10, 10);
    pointer("pointermove", window, 10, -20);
    pointer("pointermove", window, 10, -200);
    expect(document.body.querySelector(".drag-ghost")?.classList.contains("castable")).toBe(true);
    pointer("pointerup", window, 10, -200);
    slot.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    flushSync();

    expect(clicked).toEqual([]);
    expect(dragged).toHaveLength(1);
    expect(dragged[0]).toMatchObject({ zone: "command", face: undefined });
    expect(dragged[0].card.instance_id).toBe("kenrith");
    // The auto-tap preview is asked about a cast from the command zone,
    // so the tax is in what it plans to tap.
    expect(previewCalls[0]?.[1]).toBe("kenrith");
    expect(previewCalls[0]?.[2]).toMatchObject({ cast: { fromZone: "command" } });

    // PlayerPanel wires the drop as onPlayCard(card, zone, face, true),
    // which the chain turns into strict + auto_tap on the wire.
    const params: Record<string, unknown> = { instance_id: "kenrith" };
    applyCastChoices(params, castChoicesBase("command", true));
    expect(params).toEqual({
      instance_id: "kenrith",
      from_zone: "command",
      strict: true,
      auto_tap: true,
    });
  });

  it("wears the tax price tag when the price is not the printed cost", () => {
    const { container } = mountStrip(
      snap({ mine: [kenrith], moves: [castMove("kenrith")], casts: { kenrith: 1 } }),
    );
    const tag = slotOf(container, "Kenrith").querySelector(".cost-tag");
    expect(tag?.getAttribute("aria-label")).toBe("costs {6}{W} to cast from the command zone");
    expect(tag?.getAttribute("title")).toContain("commander tax +{2}");
  });

  it("the narrow-panel chip counts the commander", () => {
    const view = snap({ mine: [kenrith], exile: [exiled], moves: [castMove("kenrith")] });
    const { container } = mountStrip(view, legalActionsOf(view));
    const chip = container.querySelector(".strip-toggle");
    expect(chip?.getAttribute("aria-label")).toBe(
      "1 commander and 1 exiled card you may cast, 1 ready",
    );
    expect(chip?.querySelector(".toggle-count")?.textContent).toBe("2");
  });

  it("with only a commander, the chip names the commander alone", () => {
    const view = snap({ mine: [kenrith], moves: [castMove("kenrith")] });
    const { container } = mountStrip(view, legalActionsOf(view));
    expect(container.querySelector(".strip-toggle")?.getAttribute("aria-label")).toBe(
      "1 commander you may cast, 1 ready",
    );
  });
});
