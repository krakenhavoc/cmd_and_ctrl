// @vitest-environment jsdom
//
// ADR 0203 (#4096): a player's dungeon on the player panel. The chip
// names the dungeon and the room the venture marker is in; its card
// lists every room, the current one and the ones the next venture may
// enter marked in words; and a player who has completed a dungeon keeps
// a chip with the count.

import { describe, it, expect, afterEach } from "vitest";
import { flushSync } from "svelte";

import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { DungeonView, PlayerView } from "./protocol";
import { completedPhrase, dungeonDescription, dungeonRoomLines } from "./dungeon";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  document.querySelectorAll(".dungeon-card").forEach((n) => n.remove());
});

const undercity = (room: string): DungeonView => ({
  id: "d1",
  key: "undercity",
  name: "Undercity",
  room,
  rooms: [
    {
      key: "secret-entrance",
      name: "Secret Entrance",
      text: "Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
      leads_to: ["forge", "lost-well"],
    },
    {
      key: "forge",
      name: "Forge",
      text: "Put two +1/+1 counters on target creature.",
      leads_to: ["trap", "arena"],
    },
    { key: "lost-well", name: "Lost Well", text: "Scry 2.", leads_to: ["arena", "stash"] },
    { key: "trap", name: "Trap!", text: "Target player loses 5 life.", leads_to: ["archives"] },
    {
      key: "arena",
      name: "Arena",
      text: "Goad target creature.",
      leads_to: ["archives", "catacombs"],
    },
    { key: "stash", name: "Stash", text: "Create a Treasure token.", leads_to: ["catacombs"] },
    { key: "archives", name: "Archives", text: "Draw a card.", leads_to: ["throne"] },
    {
      key: "catacombs",
      name: "Catacombs",
      text: "Create a 4/1 black Skeleton creature token with menace.",
      leads_to: ["throne"],
    },
    {
      key: "throne",
      name: "Throne of the Dead Three",
      text: "Reveal the top ten cards of your library.",
    },
  ],
});

const seat = (dungeon?: DungeonView, completed?: string[]): PlayerView =>
  ({
    id: "alice",
    name: "Alice",
    seat: 0,
    life: 40,
    library: { kind: "library", count: 0, cards: [] },
    hand: { kind: "hand", count: 0, cards: [] },
    graveyard: { kind: "graveyard", count: 0, cards: [] },
    command: { kind: "command", count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
    dungeon,
    completed_dungeons: completed,
  }) as unknown as PlayerView;

function mountSeat(dungeon?: DungeonView, completed?: string[]) {
  return render(
    PlayerIdentity as never,
    {
      seat: seat(dungeon, completed),
      isSelf: false,
      isActive: false,
      hasPriority: false,
      attackTargetable: false,
      isMonarch: false,
      isInitiative: true,
      sendAction: () => {},
    } as never,
  );
}

const chipOf = (container: HTMLElement) =>
  container.querySelector<HTMLButtonElement>("button.dungeon");
const cardOf = () => document.querySelector<HTMLElement>(".dungeon-card");

describe("dungeon.ts", () => {
  it("marks the current room and the rooms the next venture may enter", () => {
    const lines = dungeonRoomLines(undercity("forge"));
    expect(lines.filter((l) => l.current).map((l) => l.name)).toEqual(["Forge"]);
    expect(lines.filter((l) => l.next).map((l) => l.name)).toEqual(["Trap!", "Arena"]);
    expect(lines[0].leadsTo).toEqual(["Forge", "Lost Well"]);
  });

  it("says the last room completes the dungeon", () => {
    expect(dungeonDescription(undercity("throne"), [])).toContain(
      "the next venture completes Undercity",
    );
  });

  it("counts completed dungeons in words", () => {
    expect(completedPhrase([])).toBe("");
    expect(completedPhrase(["Undercity"])).toBe("Completed Undercity");
    expect(completedPhrase(["Undercity", "Undercity"])).toBe("Completed Undercity twice");
  });
});

describe("the dungeon chip", () => {
  it("names the dungeon and the room", () => {
    const { container } = mountSeat(undercity("lost-well"));
    const chip = chipOf(container)!;
    expect(chip).not.toBeNull();
    expect(chip.querySelector(".label")?.textContent).toBe("Undercity");
    expect(chip.querySelector(".room")?.textContent).toBe("Lost Well");
    expect(chip.getAttribute("aria-label")).toBe("Undercity, in Lost Well");
    const desc = document.getElementById(chip.getAttribute("aria-describedby")!);
    expect(desc!.textContent).toContain(
      "Next: Arena (Goad target creature.) or Stash (Create a Treasure token.)",
    );
  });

  it("opens on a press with every room, the current and next ones marked in words", () => {
    const { container } = mountSeat(undercity("forge"));
    const chip = chipOf(container)!;
    chip.click();
    flushSync();
    expect(chip.getAttribute("aria-expanded")).toBe("true");
    const card = cardOf()!;
    expect(card).not.toBeNull();
    const items = [...card.querySelectorAll("li.room-line")];
    expect(items).toHaveLength(9);
    const current = card.querySelector("li.current")!;
    expect(current.textContent).toContain("Forge");
    expect(current.textContent).toContain("you are here");
    const next = [...card.querySelectorAll("li.next")];
    expect(next.map((li) => li.querySelector(".name")?.textContent)).toEqual(["Trap!", "Arena"]);
    expect(next.map((li) => li.querySelector(".when")?.textContent)).toEqual(["next", "next"]);
  });

  it("stays, with the count, once a dungeon is completed and none is owned", () => {
    const { container } = mountSeat(undefined, ["Undercity"]);
    const chip = chipOf(container)!;
    expect(chip).not.toBeNull();
    expect(chip.querySelector(".done")?.textContent).toBe("1");
    expect(chip.getAttribute("aria-label")).toBe("Completed Undercity");
  });

  it("is absent for a player with no dungeon history", () => {
    const { container } = mountSeat();
    expect(chipOf(container)).toBeNull();
  });
});
