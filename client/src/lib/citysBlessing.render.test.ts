// @vitest-environment jsdom
//
// The client half of the city's blessing (CR 702.131c, #2696). The
// blessing is a designation with no card behind it, so the marker
// beside the player's name is the only place a table can see who has it.
// It must show on a seat's own panel AND on an opponent's, and must not
// be a button: nothing toggles it.

import { describe, it, expect, afterEach } from "vitest";

import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seat = (blessed: boolean): PlayerView =>
  ({
    id: "p1",
    name: "Pat",
    seat: 0,
    life: 40,
    library: { kind: "library", count: 0, cards: [] },
    hand: { kind: "hand", count: 0, cards: [] },
    graveyard: { kind: "graveyard", count: 0, cards: [] },
    command: { kind: "command", count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
    citys_blessing: blessed || undefined,
  }) as unknown as PlayerView;

const props = (s: PlayerView, isSelf: boolean) => ({
  seat: s,
  isSelf,
  isActive: false,
  hasPriority: false,
  attackTargetable: false,
  isMonarch: false,
  isInitiative: false,
  sendAction: () => {},
});

describe("PlayerIdentity city's blessing", () => {
  it.each([
    ["an opponent", false],
    ["your own seat", true],
  ])("shows the marker on %s", (_label, isSelf) => {
    const { container } = render(PlayerIdentity as never, props(seat(true), isSelf) as never);
    const chips = container.querySelectorAll(".marker.citys-blessing");
    expect(chips).toHaveLength(1);
    expect(chips[0].getAttribute("title")).toBe("the city's blessing");
    // A designation, not a toggle.
    expect(chips[0].tagName).toBe("SPAN");
  });

  it("shows nothing for a seat without it", () => {
    const { container } = render(PlayerIdentity as never, props(seat(false), false) as never);
    expect(container.querySelectorAll(".citys-blessing")).toHaveLength(0);
  });
});
