// @vitest-environment jsdom
//
// outsidePlay.render.test.ts — #3955, Wish. "You may play a card you own
// from outside the game this turn" with the card database as the pool: the
// card is named when it is played, so the seat's sideboard pile and tab
// carry a button that sends the `outside_play` action.

import { describe, it, expect, afterEach } from "vitest";

import PileBar from "./components/board/PileBar.svelte";
import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, GameView, PlayerView, ZoneView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const zone = (kind: string, count = 0): ZoneView => ({ kind, count, cards: [] });

function seat(extras: Partial<PlayerView> = {}): PlayerView {
  return {
    id: ME,
    name: "Me",
    seat: 0,
    life: 40,
    library: zone("library", 60),
    hand: zone("hand"),
    graveyard: zone("graveyard"),
    command: zone("command"),
    commander_damage: {},
    life_history: [],
    ...extras,
  } as unknown as PlayerView;
}

function game(s: PlayerView): GameView {
  return {
    id: "g",
    state: "active",
    seats: [s],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    turn: { seq: 1, number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

const RIGHT = { source_name: "Wish", can_name: true };

describe("the Wish right's button", () => {
  it("is in the sideboard tab, and sends outside_play with no params", () => {
    const sent: { type: ActionType; params?: unknown }[] = [];
    let closed = false;
    const r = render(
      ZoneBrowserModal as never,
      {
        view: game(seat({ outside_play: RIGHT })),
        viewerID: ME,
        zoneKind: "sideboard",
        ownerSeat: { id: ME, name: "Me" },
        sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
        onClose: () => (closed = true),
      } as never,
    );
    const tabs = [...r.container.querySelectorAll<HTMLElement>('[role="tab"]')].map((t) =>
      t.textContent?.trim(),
    );
    expect(tabs, "an empty sideboard still has its tab while the right can name a card").toContain(
      "sideboard",
    );
    const button = r.container.querySelector<HTMLButtonElement>(
      '[data-testid="name-outside-card"]',
    );
    expect(button?.textContent).toContain("Name a card to play (Wish)");
    click(button!);
    expect(sent).toEqual([{ type: "outside_play", params: undefined }]);
    expect(closed).toBe(true);
  });

  it("is absent without the right, or when the sideboard casts its cards directly", () => {
    for (const extras of [{}, { outside_play: { source_name: "Wish" } }]) {
      const r = render(
        ZoneBrowserModal as never,
        {
          view: game(seat(extras)),
          viewerID: ME,
          zoneKind: "sideboard",
          ownerSeat: { id: ME, name: "Me" },
          sendAction: () => {},
          onClose: () => {},
        } as never,
      );
      expect(r.container.querySelector('[data-testid="name-outside-card"]')).toBeNull();
      cleanup();
    }
  });

  it("gives an empty sideboard its pile while the right can name a card", () => {
    const mount = (extras: Partial<PlayerView>) =>
      render(
        PileBar as never,
        { seat: seat(extras), exile: zone("exile"), isSelf: true, sendAction: () => {} } as never,
      ).container;
    expect(mount({ outside_play: RIGHT }).querySelector('[data-pile="sideboard"]')).not.toBeNull();
    cleanup();
    expect(mount({}).querySelector('[data-pile="sideboard"]')).toBeNull();
  });
});
