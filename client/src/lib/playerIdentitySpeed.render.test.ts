// @vitest-environment jsdom
//
// ADR 0138 (CR 702.179): a player's speed is public, so the seat shows
// it on every seat, its own included, as a small chip beside the other
// player markers. Nothing while a seat has no speed, and the chip is
// marked at max speed.

import { describe, it, expect, afterEach } from "vitest";

import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { PlayerView } from "./protocol";
import { speedTitle } from "./speed";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seat = (speed?: number): PlayerView =>
  ({
    id: "me",
    name: "Me",
    seat: 0,
    life: 40,
    library: { kind: "library", count: 0, cards: [] },
    hand: { kind: "hand", count: 0, cards: [] },
    graveyard: { kind: "graveyard", count: 0, cards: [] },
    command: { kind: "command", count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
    speed,
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

describe("PlayerIdentity speed", () => {
  it("shows no chip for a seat with no speed", () => {
    const { container } = render(PlayerIdentity as never, props(seat(), false) as never);
    expect(container.querySelector(".marker.speed")).toBeNull();
  });

  it("shows the speed on an opponent's seat and on your own", () => {
    for (const isSelf of [false, true]) {
      const { container } = render(PlayerIdentity as never, props(seat(2), isSelf) as never);
      const chip = container.querySelector(".marker.speed");
      expect(chip, `isSelf=${isSelf}`).not.toBeNull();
      expect(chip?.textContent).toContain("2");
      expect(chip?.classList.contains("max")).toBe(false);
      expect(chip?.getAttribute("title")).toBe(speedTitle(2));
      cleanup();
    }
  });

  it("marks max speed", () => {
    const { container } = render(PlayerIdentity as never, props(seat(4), false) as never);
    const chip = container.querySelector(".marker.speed");
    expect(chip?.classList.contains("max")).toBe(true);
    expect(chip?.getAttribute("title")).toContain("Max speed");
  });
});

describe("speedTitle", () => {
  it("says the rule below max speed and names max speed at 4", () => {
    expect(speedTitle(0)).toBe("");
    expect(speedTitle(1)).toContain("Speed 1 of 4");
    expect(speedTitle(4)).toContain("Max speed");
  });
});
