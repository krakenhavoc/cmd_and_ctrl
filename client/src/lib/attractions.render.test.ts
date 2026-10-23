// @vitest-environment jsdom
//
// attractions.render.test.ts — ADR 0207 (#2121). A seat playing with
// Attractions has a face-down Attraction deck and a face-up junkyard in
// the command zone. The rail draws each only while it holds a card: the
// deck as a count on a card back that nobody can open, the junkyard as a
// pile that opens a read-only junkyard tab of the zone browser, for every
// viewer, because the junkyard is public.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { get } from "svelte/store";

import PileBar from "./components/board/PileBar.svelte";
import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { closeZoneBrowser, zoneBrowser } from "./zoneBrowser";
import { buildMovePayload, canManageZone, cardsForZone } from "./zoneBrowser.logic";
import { L } from "./labels";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  closeZoneBrowser();
});
beforeEach(() => closeZoneBrowser());

const ME = "me";
const THEM = "them";

const zone = (kind: string, count = 0): ZoneView => ({ kind, count, cards: [] });

function attraction(id: string, name: string): CardView {
  return {
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    type_line: "Artifact — Attraction",
    attraction_lights: [2, 6],
    known_by_you: true,
  } as unknown as CardView;
}

const BOOTH = attraction("j1", "Information Booth");
const RIDE = attraction("j2", "Spinny Ride");

function seatOf(id: string, extras: Partial<PlayerView> = {}): PlayerView {
  return {
    id,
    name: id === ME ? "Me" : "Them",
    seat: id === ME ? 0 : 1,
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

function gameWith(me: Partial<PlayerView>, them: Partial<PlayerView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seatOf(ME, me), seatOf(THEM, them)],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    turn: { seq: 1, number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

function mountBar(me: Partial<PlayerView>, isSelf = true) {
  return render(
    PileBar as never,
    { seat: seatOf(ME, me), exile: zone("exile"), isSelf, sendAction: () => {} } as never,
  ).container;
}

describe("the junkyard's zone logic", () => {
  const view = gameWith({ junkyard: [BOOTH, RIDE] });

  it("reads the seat's junkyard for any viewer", () => {
    expect(cardsForZone(view, "junkyard", ME).map((c) => c.instance_id)).toEqual(["j1", "j2"]);
    expect(cardsForZone(view, "junkyard", THEM)).toEqual([]);
  });

  it("is read only, and builds no move (CR 717.6a)", () => {
    expect(canManageZone("junkyard", ME, ME)).toBe(false);
    for (const dest of ["hand", "battlefield", "library"] as const) {
      expect(buildMovePayload("junkyard", ME, ME, BOOTH, dest)).toBeNull();
    }
  });
});

describe("the player rail", () => {
  it("draws no Attraction pile for a seat that plays without them", () => {
    const c = mountBar({});
    expect(c.querySelector('[data-pile="attraction-deck"]')).toBeNull();
    expect(c.querySelector('[data-pile="junkyard"]')).toBeNull();
    expect(c.textContent).not.toContain(L.attractionDeck);
    expect(c.textContent).not.toContain(L.junkyard);
  });

  it("shows the Attraction deck as a face-down count nobody can open", () => {
    for (const isSelf of [true, false]) {
      const c = mountBar({ attraction_deck_count: 8 }, isSelf);
      const pile = c.querySelector<HTMLButtonElement>('[data-pile="attraction-deck"]');
      expect(pile, `isSelf=${isSelf}`).not.toBeNull();
      expect(pile!.textContent).toContain(L.attractionDeck);
      expect(pile!.textContent).toContain("8");
      expect(pile!.getAttribute("aria-label")).toContain(`${L.attractionDeck}: 8 cards`);
      expect(pile!.disabled).toBe(true);
      // A back, never a card face: the deck's cards are on nobody's wire.
      expect(pile!.querySelector(".card:not(.face-down) img")).toBeNull();
      cleanup();
    }
  });

  it("draws no Attraction deck once it is empty", () => {
    const c = mountBar({ attraction_deck_count: 0, junkyard: [BOOTH] });
    expect(c.querySelector('[data-pile="attraction-deck"]')).toBeNull();
    expect(c.querySelector('[data-pile="junkyard"]')).not.toBeNull();
  });

  it("shows the junkyard on any seat, and opens its zone browser tab", () => {
    const c = mountBar({ junkyard: [BOOTH, RIDE] }, false);
    const pile = c.querySelector<HTMLButtonElement>('[data-pile="junkyard"]');
    expect(pile).not.toBeNull();
    expect(pile!.getAttribute("aria-label")).toContain(`${L.junkyard}: 2 cards`);
    expect(pile!.disabled).toBe(false);
    pile!.click();
    expect(get(zoneBrowser)).toEqual({ zoneKind: "junkyard", ownerID: ME, ownerName: "Me" });
  });
});

describe("the zone browser's junkyard tab", () => {
  function mountModal(view: GameView, viewerID: string | null, zoneKind: "junkyard" | "graveyard") {
    const sent: ActionType[] = [];
    const r = render(
      ZoneBrowserModal as never,
      {
        view,
        viewerID,
        zoneKind,
        ownerSeat: { id: ME, name: "Me" },
        sendAction: (t: ActionType) => sent.push(t),
        onClose: () => {},
        onCastCard: () => {},
      } as never,
    );
    return { c: r.container, sent };
  }
  const tabsOf = (c: HTMLElement) =>
    [...c.querySelectorAll<HTMLElement>('[role="tab"]')].map((t) => t.textContent?.trim());

  it("lists the junkyard, says it is read only, and moves nothing", () => {
    const m = mountModal(gameWith({ junkyard: [BOOTH, RIDE] }), THEM, "junkyard");
    const text = m.c.textContent ?? "";
    expect(text).toContain("junkyard");
    expect(text).toContain("read only");
    expect(text).toContain("Information Booth");
    expect(text).toContain("Spinny Ride");
    expect(m.c.querySelectorAll(".cell").length).toBe(2);
    expect(m.c.querySelector("button.move, [data-move]")).toBeNull();
    expect(m.sent).toEqual([]);
  });

  it("is a tab for every viewer while the seat has a junkyard, and only then", () => {
    for (const viewer of [ME, THEM, null]) {
      expect(tabsOf(mountModal(gameWith({ junkyard: [BOOTH] }), viewer, "graveyard").c)).toContain(
        "junkyard",
      );
      cleanup();
    }
    expect(tabsOf(mountModal(gameWith({}), ME, "graveyard").c)).not.toContain("junkyard");
  });
});
