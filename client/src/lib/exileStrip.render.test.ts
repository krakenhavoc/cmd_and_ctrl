// @vitest-environment jsdom
//
// exileStrip.render.test.ts — #1389. What the strip DOES with a click,
// which no pure function sits between: a castable card hands itself to
// the Board's one cast chain with the zone and the grant's face (the
// #874 contract the zone browser's exile button keeps), and a card that
// is not castable now does nothing. Plus the tag and the caption, as
// markup.

import { describe, it, expect, afterEach } from "vitest";

import ExileStrip from "./components/board/ExileStrip.svelte";
import type { CardView, GameView, LegalMoveView } from "./protocol";
import type { CastSourceZone } from "./targeting";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const THEM = "them";

function snap(cards: CardView[], moves?: LegalMoveView[]): GameView {
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
  const handed: Handed[] = [];
  const r = render(
    ExileStrip as never,
    {
      view,
      viewerID: ME,
      onCastCard: (card: CardView, zone: CastSourceZone, face?: number) =>
        handed.push({ card, zone, face }),
    } as never,
  );
  return { container: r.container, handed, r };
}

const cardEl = (c: HTMLElement, name: string) =>
  c.querySelector<HTMLElement>(`.card[aria-label='${name}']`);

const plotted: CardView = {
  instance_id: "plotted",
  name: "Plotted Bolt",
  owner: ME,
  controller: ME,
  type_line: "Instant",
  mana_cost: "{R}",
  exile_play: { player: ME, cast_only: true, cost_override: "{0}" },
  castable_here: true,
  cast_prices: [{ cost: "{0}" }],
};

describe("ExileStrip — #1389", () => {
  it("renders nothing when the viewer may cast nothing from exile", () => {
    const { container } = mount(snap([]));
    expect(container.querySelector(".exile-strip")).toBeNull();
  });

  it("hands a castable card to the cast chain with its zone and face", () => {
    const siege: CardView = {
      instance_id: "siege",
      name: "Invasion",
      owner: THEM,
      controller: THEM,
      type_line: "Battle — Siege",
      exile_play: { player: ME, faces: [1], cast_only: true },
      castable_here: true,
      cast_prices: [{ cost: "{0}" }],
    };
    const { container, handed } = mount(snap([siege], [castMove("siege")]));
    click(cardEl(container, "Invasion")!);
    expect(handed).toHaveLength(1);
    expect(handed[0].zone).toBe("exile");
    expect(handed[0].face).toBe(1);
    expect(handed[0].card.instance_id).toBe("siege");
  });

  it("does nothing on a card the server's move list does not offer — greyed like the hand", () => {
    const { container, handed } = mount(snap([plotted], [castMove("something-else")]));
    click(cardEl(container, "Plotted Bolt")!);
    expect(handed).toEqual([]);
    expect(container.querySelector(".strip-slot.blocked")).not.toBeNull();
  });

  it("draws the price tag only when the price is not the printed cost", () => {
    const { container } = mount(snap([plotted], [castMove("plotted")]));
    const tag = container.querySelector(".cost-tag");
    expect(tag?.getAttribute("aria-label")).toBe("costs free to cast from exile");
    expect(tag?.textContent?.replace(/\s/g, "")).toBe("0");

    cleanup();
    const printed = mount(
      snap([{ ...plotted, cast_prices: [{ cost: "{R}", printed: true }] }], [castMove("plotted")]),
    );
    expect(printed.container.querySelector(".cost-tag")).toBeNull();
  });

  it("dims a card waiting on a later turn and says when", () => {
    const warped: CardView = {
      instance_id: "warped",
      name: "Warped Drake",
      owner: ME,
      controller: ME,
      type_line: "Creature — Drake",
      exile_play: { player: ME, cast_only: true, not_before_seq: 5 },
    };
    const { container, handed } = mount(snap([warped]));
    expect(container.querySelector(".strip-slot.later")).not.toBeNull();
    expect(container.querySelector(".hint")?.textContent).toBe("next turn");
    click(cardEl(container, "Warped Drake")!);
    expect(handed).toEqual([]);
  });

  it("drops a card from the strip when the next frame no longer carries it", () => {
    const { container, r } = mount(snap([plotted], [castMove("plotted")]));
    expect(cardEl(container, "Plotted Bolt")).not.toBeNull();
    r.setProps({ view: snap([]) } as never);
    expect(container.querySelector(".exile-strip")).toBeNull();
  });

  it("offers the collapsed chip with the count and how many are ready", () => {
    const { container } = mount(snap([plotted], [castMove("plotted")]));
    const chip = container.querySelector(".strip-toggle");
    expect(chip?.getAttribute("aria-label")).toBe("1 exiled card you may cast, 1 ready");
    click(chip!);
    expect(container.querySelector(".exile-strip.open")).not.toBeNull();
  });
});
