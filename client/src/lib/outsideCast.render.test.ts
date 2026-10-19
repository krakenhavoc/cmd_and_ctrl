// @vitest-environment jsdom
//
// outsideCast.render.test.ts — #3955. A card the viewer owns outside the
// game that a permission lets them cast or play carries a button in the
// outside-the-game view (#3833: no longer a sideboard tab), and clicking it hands the card to the Board's cast chain
// with the `outside` source zone.

import { describe, it, expect, afterEach } from "vitest";

import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import type { CastSourceZone } from "./targeting";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const zone = (kind: string, count = 0): ZoneView => ({ kind, count, cards: [] });

function sideCard(id: string, name: string, over: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    type_line: "Creature — Eldrazi",
    known_by_you: true,
    ...over,
  } as unknown as CardView;
}

function gameWith(sideboard: CardView[]): GameView {
  const seat = {
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
    sideboard,
    sideboard_count: sideboard.length,
  } as unknown as PlayerView;
  return {
    id: "g",
    state: "active",
    seats: [seat],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    turn: { seq: 1, number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

function mount(sideboard: CardView[]) {
  const cast: { id: string; from: CastSourceZone }[] = [];
  const r = render(
    ZoneBrowserModal as never,
    {
      view: gameWith(sideboard),
      viewerID: ME,
      zoneKind: "sideboard",
      ownerSeat: { id: ME, name: "Me" },
      sendAction: (_t: ActionType) => {},
      onClose: () => {},
      onCastCard: (card: CardView, from: CastSourceZone) =>
        cast.push({ id: card.instance_id, from }),
    } as never,
  );
  return { c: r.container, cast };
}

describe("the outside-the-game view's cast button", () => {
  it("is on a granted card only, and casts it from outside the game", () => {
    const granted = sideCard("s1", "Ulamog's Crusher", {
      castable_here: true,
      exile_play: { player: ME, cost_override: "{0}", cast_only: true } as CardView["exile_play"],
    });
    const plain = sideCard("s2", "Emrakul's Hatcher");
    const m = mount([granted, plain]);
    const buttons = m.c.querySelectorAll<HTMLButtonElement>(".actions.always button");
    expect(buttons.length, "one cast button, on the granted card").toBe(1);
    expect(buttons[0].textContent?.trim()).toBe("cast free");
    expect(buttons[0].getAttribute("aria-label")).toContain("from outside the game");
    click(buttons[0]);
    expect(m.cast).toEqual([{ id: "s1", from: "outside" }]);
  });

  it("says play for a land, and for a card played at its own cost", () => {
    const land = sideCard("s1", "Forest", {
      type_line: "Basic Land — Forest",
      castable_here: true,
    });
    const bear = sideCard("s2", "Grizzly Bears", {
      castable_here: true,
      exile_play: { player: ME } as CardView["exile_play"],
    });
    const m = mount([land, bear]);
    const labels = [...m.c.querySelectorAll<HTMLButtonElement>(".actions.always button")].map((b) =>
      b.textContent?.trim(),
    );
    expect(labels).toEqual(["play", "play"]);
  });

  it("offers nothing on a sideboard no permission opens", () => {
    const m = mount([sideCard("s1", "Ulamog's Crusher")]);
    expect(m.c.querySelectorAll(".actions.always button").length).toBe(0);
    // #3833: nor does the view list it: the sideboard is not on the table.
    expect(m.c.querySelectorAll(".cell").length).toBe(0);
    expect(m.cast).toEqual([]);
  });
});
