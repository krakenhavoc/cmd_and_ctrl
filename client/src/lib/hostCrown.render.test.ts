// @vitest-environment jsdom
//
// The host's crown on a seat (ADR 0075 §2.1).
//
// Hosting is a seat ROLE — who may change the house rules and, when
// the table allows it, spawn — and the table can only hold one host
// accountable if everyone can see which seat it is. The lobby chip
// shipped with sub-PR 2; this is its twin at the table.
//
// It is asserted as a chip beside the name rather than as "a crown",
// because the crown glyph already means MONARCH in the marker column
// two rows down. Two crowns on one seat meaning two different things
// is the confusion this markup is arranged to avoid, and a test that
// only counted crowns would not notice if they merged.

import { describe, it, expect, afterEach } from "vitest";

import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seat = (over: Partial<PlayerView>): PlayerView =>
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
    ...over,
  }) as unknown as PlayerView;

const props = (s: PlayerView, over: Record<string, unknown> = {}) => ({
  seat: s,
  isSelf: false,
  isActive: false,
  hasPriority: false,
  attackTargetable: false,
  isMonarch: false,
  isInitiative: false,
  sendAction: () => {},
  ...over,
});

function host(container: HTMLElement) {
  return container.querySelector(".tag.host");
}

describe("the host crown", () => {
  it("is absent on an ordinary seat", () => {
    const { container } = render(PlayerIdentity as never, props(seat({})) as never);
    expect(host(container)).toBeNull();
  });

  it("marks the host's seat, and says what hosting means", () => {
    const { container } = render(PlayerIdentity as never, props(seat({ is_host: true })) as never);
    const chip = host(container);
    expect(chip).not.toBeNull();
    expect(chip?.textContent).toContain("host");
    expect(chip?.getAttribute("title")).toMatch(/house rules/i);
  });

  it("does not depend on the viewer being that seat", () => {
    // Everyone sees the host, including a spectator: it is the answer
    // to "who do I ask about the undo limit".
    const { container } = render(
      PlayerIdentity as never,
      props(seat({ is_host: true }), { isSelf: false }) as never,
    );
    expect(host(container)).not.toBeNull();
  });

  it("stays distinct from the monarch's crown", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat({ is_host: true }), { isMonarch: true, isSelf: true }) as never,
    );
    // Both marks are present and neither is the other's element.
    const hostChip = host(container);
    const monarch = container.querySelector(".marker.monarch");
    expect(hostChip).not.toBeNull();
    expect(monarch).not.toBeNull();
    expect(hostChip).not.toBe(monarch);
  });
});
