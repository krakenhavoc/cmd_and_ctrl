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
import CommandStrip from "./components/board/CommandStrip.svelte";
import { zoneBrowser } from "./zoneBrowser";
import { get } from "svelte/store";
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

  // #2349: the strip is the seat's command zone on the board now, so its
  // commanders' group carries the tile's name.
  it("groups the commanders under the command zone's name", () => {
    const { container } = mountStrip(
      snap({ mine: [kenrith], exile: [exiled], moves: [castMove("kenrith")] }),
    );
    const group = container.querySelector("[role='group'][aria-label='Me command zone, 1 card']");
    expect(group).not.toBeNull();
    expect(group?.querySelector(".card[aria-label='Kenrith']")).not.toBeNull();
    expect(group?.querySelector(".card[aria-label='Impulse Bolt']")).toBeNull();
  });

  it("a click starts the cast chain from the command zone", () => {
    const view = snap({ mine: [kenrith], moves: [castMove("kenrith")] });
    const strip = mountStrip(view);
    click(cardEl(strip.container, "Kenrith")!);
    expect(strip.clicked).toHaveLength(1);
    expect(strip.clicked[0].card.instance_id).toBe("kenrith");
    expect(strip.clicked[0].zone).toBe("command");
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

// #2349: another seat's commander, beside that seat's hand, where the command
// zone tile in the rail used to be the only place it was face up.
describe("another seat's command strip — #2349", () => {
  const seat = (casts?: Record<string, number>) => ({
    id: THEM,
    name: "Them",
    command: { kind: "command", count: 1, cards: [theirCommander] },
    commander_casts: casts,
  });

  it("shows their commander face up, under the command zone's name", () => {
    const { container } = render(CommandStrip as never, { seat: seat() } as never);
    const group = container.querySelector("[role='group'][aria-label='Them command zone, 1 card']");
    expect(group?.querySelector(".card[aria-label='Their Commander']")).not.toBeNull();
    expect(container.querySelector(".tax-badge")).toBeNull();
  });

  it("wears the commander tax, read off commander_casts", () => {
    const { container } = render(CommandStrip as never, { seat: seat({ theirs: 2 }) } as never);
    expect(container.querySelector(".tax-badge")?.textContent).toBe("+4");
  });

  it("a click opens their command zone in the browser, and casts nothing", () => {
    const { container } = render(CommandStrip as never, { seat: seat() } as never);
    click(cardEl(container, "Their Commander")!);
    expect(get(zoneBrowser)).toMatchObject({
      zoneKind: "command",
      ownerID: THEM,
      ownerName: "Them",
    });
  });

  it("renders nothing for an empty command zone", () => {
    const empty = { id: THEM, name: "Them", command: { kind: "command", count: 0, cards: [] } };
    const { container } = render(CommandStrip as never, { seat: empty } as never);
    expect(container.querySelector(".command-strip")).toBeNull();
  });
});
