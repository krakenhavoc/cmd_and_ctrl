// @vitest-environment jsdom
//
// ADR 0216 §10: the face-up plane beside the table. It draws the plane
// with its name, the deck's count under its contract label and the
// planar controller, and nothing at a table that does not play
// Planechase.

import { afterEach, describe, expect, it } from "vitest";

import PlanarZone from "./components/board/PlanarZone.svelte";
import { L } from "./labels";
import type { GameView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

afterEach(cleanup);

function view(planechase?: GameView["planechase"]): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      { id: "alice", seat: 0, name: "Alice" },
      { id: "bob", seat: 1, name: "Bob" },
    ],
    turn: { active_seat: 1, priority_holder: 1, step: "precombat_main", number: 2 },
    battlefield: { count: 0, cards: [] },
    stack: { count: 0, cards: [] },
    exile: { count: 0, cards: [] },
    planechase,
  } as unknown as GameView;
}

const PLANE = {
  instance_id: "p1",
  name: "Fixture Plane",
  owner: "bob",
  controller: "bob",
  type_line: "Plane — Fixtureworld",
  layout: "planar",
};

describe("PlanarZone", () => {
  it("draws nothing at a table without Planechase", () => {
    const { container } = render(PlanarZone as never, { view: view() } as never);
    expect(container.querySelector('[role="group"]')).toBeNull();
  });

  it("draws the face-up plane, the deck count and the planar controller", () => {
    const { container } = render(
      PlanarZone as never,
      {
        view: view({
          planar_controller: "bob",
          deck_count: 73,
          face_up: [PLANE as never],
          die_rolls: 0,
          die_cost: "{0}",
        }),
      } as never,
    );
    const zone = container.querySelector(`[aria-label="${L.planarZone}"]`);
    expect(zone).not.toBeNull();
    expect(zone!.textContent).toContain("Fixture Plane");
    expect(zone!.textContent).toContain("Bob");
    const deck = container.querySelector(`[aria-label="${L.planarDeck(73)}"]`);
    expect(deck?.textContent).toContain("73 planes");
  });

  it("says when the first plane is still to come", () => {
    const { container } = render(
      PlanarZone as never,
      { view: view({ deck_count: 74, face_up: [], die_rolls: 0, die_cost: "{0}" }) } as never,
    );
    expect(container.textContent).toContain("after the mulligans");
  });
});
