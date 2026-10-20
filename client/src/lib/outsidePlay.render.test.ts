// @vitest-environment jsdom
//
// outsidePlay.render.test.ts — #3955, Wish; the pile left the table by #3833
// and #4054. "You may play a card you own from outside the game this turn"
// with the card database as the pool: the card is named when it is played, so
// the right puts a "name a card to play" chip in the action dock, a shortcut
// that sits with the verbs of the turn. Pressing it sends the `outside_play`
// action with no params; the rest of the right is the cards it opens, behind
// a second chip that opens the read-only Sideboard tab of the zone browser,
// where each such card has its cast button.

import { describe, it, expect, afterEach } from "vitest";

import ActionDock from "./components/board/ActionDock.svelte";
import type { CardView, GameView, PlayerView } from "./protocol";
import { outsideCastable, outsideChips, outsideDock } from "./outsideRight";
import { render, click, cleanup } from "./test/render.svelte";
import { dockTestView } from "./test/dockView";

afterEach(cleanup);

const ME = "me";
const RIGHT = { source_name: "Wish", can_name: true };

function viewWith(me: Partial<PlayerView>, others: Partial<PlayerView> = {}): GameView {
  const v = dockTestView();
  v.seats = [
    { ...v.seats[0], ...me },
    { ...v.seats[1], ...others },
  ] as PlayerView[];
  return v;
}

const granted = (id: string): CardView =>
  ({
    instance_id: id,
    name: "Ulamog's Crusher",
    owner: ME,
    controller: ME,
    known_by_you: true,
    castable_here: true,
  }) as unknown as CardView;

function mountDock(view: GameView, calls: { name: number; open: number }) {
  const outside = outsideDock(view, ME, {
    name: () => calls.name++,
    open: () => calls.open++,
  });
  return render(
    ActionDock as never,
    {
      view,
      viewerHasPriority: false,
      viewerIsActive: false,
      autopassEnabled: false,
      onPassPriority: () => {},
      onPassTurn: () => {},
      onToggleAutopass: () => {},
      outside,
    } as never,
  ).container;
}

describe("the Wish right's button", () => {
  it("is in the action dock, and pressing it names a card", () => {
    const calls = { name: 0, open: 0 };
    const c = mountDock(viewWith({ outside_play: RIGHT }), calls);
    const button = c.querySelector<HTMLButtonElement>('[data-testid="name-outside-card"]');
    expect(button?.textContent?.trim()).toBe("name a card to play");
    expect(button?.getAttribute("title")).toContain("Wish");
    click(button!);
    expect(calls).toEqual({ name: 1, open: 0 });
  });

  it("is absent without the right, or when the sideboard casts its cards directly", () => {
    for (const me of [{}, { outside_play: { source_name: "Wish" } }]) {
      const c = mountDock(viewWith(me), { name: 0, open: 0 });
      expect(c.querySelector('[data-testid="name-outside-card"]')).toBeNull();
      expect(c.querySelector('[data-testid="open-outside-cards"]')).toBeNull();
      cleanup();
    }
  });

  it("sends outside_play with no params through the dock prop", () => {
    // Game.svelte maps `name` to this action; the chip carries no payload.
    expect(outsideChips(viewWith({ outside_play: RIGHT }), ME)).toEqual([
      expect.objectContaining({ kind: "name", label: "name a card to play" }),
    ]);
  });
});

describe("the cards a right opens", () => {
  it("get a second button, counted, that opens the sideboard tab", () => {
    const calls = { name: 0, open: 0 };
    const c = mountDock(
      viewWith({ outside_play: { source_name: "Spawnsire of Ulamog" }, sideboard: [granted("a")] }),
      calls,
    );
    const open = c.querySelector<HTMLButtonElement>('[data-testid="open-outside-cards"]');
    expect(open?.textContent?.trim()).toBe("outside the game (1)");
    expect(open?.getAttribute("title")).toContain("Spawnsire of Ulamog");
    click(open!);
    expect(calls).toEqual({ name: 0, open: 1 });
  });

  it("show both buttons when a named card is minted but not yet played", () => {
    const view = viewWith({ outside_play: RIGHT, sideboard: [granted("a"), granted("b")] });
    expect(outsideChips(view, ME).map((c) => c.kind)).toEqual(["name", "open"]);
    expect(outsideCastable(view, ME).length).toBe(2);
  });

  it("are read from the viewer's own seat only", () => {
    const view = viewWith({}, { sideboard: [granted("x")], outside_play: RIGHT });
    expect(outsideChips(view, ME)).toEqual([]);
    expect(outsideChips(view, null)).toEqual([]);
    expect(outsideDock(view, ME, { name: () => {}, open: () => {} })).toBeUndefined();
  });
});
