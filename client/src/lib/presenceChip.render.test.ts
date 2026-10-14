// @vitest-environment jsdom
//
// ADR 0177 §1 on the seat: who is gone, for how long, and a seat a bot
// is playing for its person. A seat holding priority while offline
// reads "offline", never "considering…".

import { describe, it, expect, afterEach } from "vitest";

import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ago = (mins: number) => new Date(Date.now() - mins * 60000 - 1000).toISOString();

const seat = (over: Partial<PlayerView>): PlayerView =>
  ({
    id: "bo",
    name: "Bo",
    seat: 1,
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

describe("the presence chip", () => {
  it("is absent on a present seat", () => {
    const { container } = render(PlayerIdentity as never, props(seat({})) as never);
    expect(container.querySelector(".tag.presence")).toBeNull();
  });

  it("says offline and for how long, and replaces considering", () => {
    const s = seat({ presence: { state: "offline", since: ago(2) } });
    const { container } = render(
      PlayerIdentity as never,
      props(s, { hasPriority: true, considering: true }) as never,
    );
    expect(container.querySelector(".tag.presence.offline")?.textContent).toBe("offline 2m");
    expect(container.querySelector(".tag.considering")).toBeNull();
  });

  it("says away for a tab hidden past two minutes", () => {
    const s = seat({ presence: { state: "hidden", since: ago(4) } });
    const { container } = render(PlayerIdentity as never, props(s) as never);
    expect(container.querySelector(".tag.presence.away")?.textContent).toBe("away 4m");
  });

  it("marks a seat a bot plays for its person", () => {
    const s = seat({ stand_in: { tier: "heuristic", since: ago(5) } });
    const { container } = render(PlayerIdentity as never, props(s) as never);
    const tag = container.querySelector(".tag.stand-in");
    expect(tag?.textContent).toBe("(bot)");
    expect(tag?.getAttribute("title")).toMatch(/A bot has been playing this seat since/);
  });
});
