// @vitest-environment jsdom
//
// phaseDisplay.dayNight.render.test.ts — ADR 0132 (#2561). The dock's
// header shows the game's day/night designation (CR 731) as one chip
// beside the turn line: nothing while the game has neither, "Day" or
// "Night" once it has one, the same for every viewer.

import { describe, it, expect, afterEach } from "vitest";

import PhaseDisplay from "./components/board/PhaseDisplay.svelte";
import { dayNightChip } from "./dayNight";
import type { PlayerView, TurnView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
] as unknown as PlayerView[];

const turn: TurnView = {
  seq: 5,
  number: 3,
  active_seat: 0,
  priority_holder: 0,
  phase: "precombat_main",
  step: "precombat_main",
};

function mount(dayNight?: string) {
  return render(PhaseDisplay as never, { turn, seats, mulligansOpen: false, dayNight } as never)
    .container;
}

describe("PhaseDisplay day and night", () => {
  it("shows no chip while the game is neither day nor night", () => {
    expect(mount().querySelector(".day-night")).toBeNull();
    expect(mount(undefined).querySelector(".day-night")).toBeNull();
  });

  it("names day", () => {
    const chip = mount("day").querySelector(".day-night");
    expect(chip?.textContent?.replace(/\s+/g, " ").trim()).toBe("☀ Day");
    expect(chip?.getAttribute("data-day-night")).toBe("day");
    expect(chip?.getAttribute("title")).toMatch(/cast no spells/);
  });

  it("names night", () => {
    const chip = mount("night").querySelector(".day-night");
    expect(chip?.textContent?.replace(/\s+/g, " ").trim()).toBe("☾ Night");
    expect(chip?.getAttribute("title")).toMatch(/two or more spells/);
  });

  it("keeps the turn line intact beside the chip", () => {
    const c = mount("night");
    expect(c.querySelector(".turn-no")?.textContent).toBe("T3");
    expect(c.querySelector(".active-name")?.textContent).toBe("Alice");
  });
});

describe("dayNightChip", () => {
  it("is null for neither and for anything it does not know", () => {
    expect(dayNightChip(undefined)).toBeNull();
    expect(dayNightChip("")).toBeNull();
    expect(dayNightChip("dusk")).toBeNull();
  });
});
