// @vitest-environment jsdom
//
// The client half of CR 114 (#623). An emblem is the one board state
// that is neither a card nor a counter: it has no art, no P/T and no
// pile to sit in, so the only place a player can learn they have one
// is this chip. Two things have to be in the markup and neither can
// be checked without a DOM — that the chip renders at all, and that
// the printed ability reaches the hover, because the label alone
// ("Elspeth, Sun's Champion emblem") does not say what the emblem
// does.

import { describe, it, expect, afterEach } from "vitest";

import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seat = (emblems?: PlayerView["emblems"]): PlayerView =>
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
    emblems,
  }) as unknown as PlayerView;

const props = (s: PlayerView) => ({
  seat: s,
  isSelf: false,
  isActive: false,
  hasPriority: false,
  attackTargetable: false,
  isMonarch: false,
  isInitiative: false,
  sendAction: () => {},
});

describe("PlayerIdentity emblems", () => {
  it("renders a chip per emblem with the printed text as the hover", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(
        seat([
          {
            instance_id: "e1",
            label: "Elspeth, Sun's Champion emblem",
            text: "Creatures you control get +2/+2 and have flying.",
          },
        ]),
      ) as never,
    );

    const chips = container.querySelectorAll(".emblem");
    expect(chips).toHaveLength(1);
    expect(chips[0].textContent).toContain("Elspeth, Sun's Champion emblem");
    // The label names it; the title is the only place the ability is.
    expect(chips[0].getAttribute("title")).toContain(
      "Creatures you control get +2/+2 and have flying.",
    );
  });

  it("renders two chips for two emblems", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(
        seat([
          { instance_id: "e1", label: "First emblem", text: "Does one thing." },
          { instance_id: "e2", label: "Second emblem", text: "Does another." },
        ]),
      ) as never,
    );
    expect(container.querySelectorAll(".emblem")).toHaveLength(2);
  });

  it("renders nothing for a seat with no emblems", () => {
    const { container } = render(PlayerIdentity as never, props(seat()) as never);
    expect(container.querySelectorAll(".emblem")).toHaveLength(0);
    expect(container.querySelectorAll(".emblems")).toHaveLength(0);
  });
});
