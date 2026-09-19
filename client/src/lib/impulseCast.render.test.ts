// @vitest-environment jsdom
//
// impulseCast.render.test.ts — #874. The exile-zone impulse button
// used to dispatch `cast_spell` straight out of ZoneBrowserModal, so
// the announcement it produced was always the empty one: a grant that
// lets you cast for the PRINTED cost could only ever announce X = 0,
// and a spell with targets or modes got none. Everything the player
// was entitled to choose lives in the Board's cast chain, and this is
// the click that has to reach it.
//
// A render test rather than a pure one because the thing that was
// wrong is what the BUTTON does, and there is no pure function
// between the click and the dispatch to examine. What the chain then
// does with the card is pinned in impulseCast.test.ts.

import { describe, it, expect, afterEach } from "vitest";

import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, ExilePlayView, GameView } from "./protocol";
import type { CastSourceZone } from "./targeting";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const THEM = "them";

function exiled(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "impulsed",
    name: "Fireball",
    owner: THEM,
    controller: THEM,
    type_line: "Sorcery",
    mana_cost: "{X}{R}",
    ...extras,
  };
}

function snapWith(cards: CardView[]): GameView {
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
    turn: { number: 7, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

interface Sent {
  type: ActionType;
  params?: unknown;
}
interface HandedUp {
  card: CardView;
  zone: CastSourceZone;
  face?: number;
}

function mount(cards: CardView[], withHandler = true) {
  const sent: Sent[] = [];
  const handed: HandedUp[] = [];
  const view = render(
    ZoneBrowserModal as never,
    {
      view: snapWith(cards),
      viewerID: ME,
      zoneKind: "exile",
      ownerSeat: { id: THEM, name: "Them" },
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      onClose: () => {},
      onCastCard: withHandler
        ? (card: CardView, zone: CastSourceZone, face?: number) => handed.push({ card, zone, face })
        : undefined,
    } as never,
  );
  return { container: view.container, sent, handed };
}

const impulseButton = (c: HTMLElement): HTMLElement | null =>
  c.querySelector("[aria-label='play from exile'] button");

const grant = (extras: Partial<ExilePlayView> = {}): ExilePlayView => ({
  player: ME,
  ...extras,
});

describe("the exile impulse button — #874", () => {
  it("hands the cast to the Board's chain instead of dispatching one", () => {
    const { container, sent, handed } = mount([exiled({ exile_play: grant() })]);
    const button = impulseButton(container);
    expect(button).not.toBeNull();
    click(button!);

    // The whole bug in one assertion: nothing goes on the wire from
    // here. The announcement is the chain's to make, once the player
    // has answered the prompts the card owes them.
    expect(sent).toEqual([]);
    expect(handed).toHaveLength(1);
    expect(handed[0].zone).toBe("exile");
    expect(handed[0].card.instance_id).toBe("impulsed");
  });

  it("passes the face a grant names, and none when it names none", () => {
    // S32: a defeated Siege sits in exile battle-side-up and its grant
    // opens face 1. The chain needs the index so every prompt after it
    // reads the half being cast.
    const siege = exiled({
      instance_id: "siege",
      name: "Invasion of New Phyrexia",
      type_line: "Battle — Siege",
      exile_play: grant({ faces: [1], cast_only: true, cost_override: "{0}" }),
    });
    const { container, handed } = mount([siege]);
    click(impulseButton(container)!);
    expect(handed[0].face).toBe(1);

    cleanup();
    const plain = mount([exiled({ exile_play: grant() })]);
    click(impulseButton(plain.container)!);
    expect(plain.handed[0].face).toBeUndefined();
  });

  it("passes face 0 when CR 715.4's Adventure grant names the creature", () => {
    // The case a bare `face?: number` could not carry: the creature
    // half of an adventure card IS face 0, so an absent field and the
    // real answer looked identical and the Board re-opened its face
    // picker on a cast with exactly one legal half (#719).
    const knight = exiled({
      instance_id: "knight",
      name: "Foulmire Knight",
      type_line: "Creature — Zombie Knight",
      mana_cost: "{B}",
      layout: "adventure",
      active_face: 0,
      faces: [
        { name: "Foulmire Knight", type_line: "Creature — Zombie Knight", mana_cost: "{B}" },
        { name: "Profane Insight", type_line: "Instant — Adventure", mana_cost: "{2}{B}" },
      ],
      exile_play: grant({ faces: [0], cast_only: true }),
    });
    const { container, handed } = mount([knight]);
    click(impulseButton(container)!);
    expect(handed[0].face).toBe(0);
    expect(handed[0].zone).toBe("exile");
  });

  it("shows no button at all when there is nothing behind it", () => {
    const { container } = mount([exiled({ exile_play: grant() })], false);
    expect(impulseButton(container)).toBeNull();
  });

  it("still shows no button for a card whose grant names somebody else", () => {
    const { container } = mount([exiled({ exile_play: grant({ player: THEM }) })]);
    expect(impulseButton(container)).toBeNull();
  });
});
