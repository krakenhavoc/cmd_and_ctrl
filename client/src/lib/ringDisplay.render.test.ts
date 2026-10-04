// @vitest-environment jsdom
//
// ADR 0114 owner decision 1 (#2076): what the table shows of the Ring.
//
//   - The Ring chip on the player panel: the emblem chip with its level
//     as a pip ("The Ring · 3"), an accessible name that says what the
//     number counts, and a card listing all four lines — the gained
//     ones in full, the rest dimmed AND marked "after the Nth
//     temptation", so the difference is in the words, not only the
//     colour.
//   - The Ring-bearer marker on the card: its own marker, apart from
//     the designation slot, because a Ring-bearer can also be monstrous
//     or harnessed and the two must both show.

import { describe, it, expect, afterEach } from "vitest";
import { flushSync } from "svelte";

import Card from "./components/board/Card.svelte";
import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { CardView, EmblemView, PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  // The chip's card is portalled to <body>.
  document.querySelectorAll(".emblem-card").forEach((n) => n.remove());
});

const LINES = [
  {
    text: "Your Ring-bearer is legendary and can't be blocked by creatures with greater power.",
    at: 1,
  },
  { text: "Whenever your Ring-bearer attacks, draw a card, then discard a card.", at: 2 },
  {
    text: "Whenever your Ring-bearer becomes blocked by a creature, the blocking creature's controller sacrifices it at end of combat.",
    at: 3,
  },
  {
    text: "Whenever your Ring-bearer deals combat damage to a player, each opponent loses 3 life.",
    at: 4,
  },
];

const ring = (level: number): EmblemView => ({
  instance_id: "ring",
  label: "The Ring",
  text: LINES.filter((l) => l.at <= level)
    .map((l) => l.text)
    .join(" "),
  level,
  lines: LINES,
});

const seat = (emblems: EmblemView[]): PlayerView =>
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
    emblems,
  }) as unknown as PlayerView;

function mountSeat(emblems: EmblemView[]) {
  return render(
    PlayerIdentity as never,
    {
      seat: seat(emblems),
      isSelf: false,
      isActive: false,
      hasPriority: false,
      attackTargetable: false,
      isMonarch: false,
      isInitiative: false,
      sendAction: () => {},
    } as never,
  );
}

const chipOf = (container: HTMLElement) =>
  container.querySelector<HTMLButtonElement>("button.emblem.levelled");
const cardOf = () => document.querySelector<HTMLElement>(".emblem-card");

describe("the Ring chip", () => {
  it("shows the label and the level as a pip, and names both", () => {
    const { container } = mountSeat([ring(3)]);
    const chip = chipOf(container)!;
    expect(chip).not.toBeNull();
    expect(chip.querySelector(".label")?.textContent).toBe("The Ring");
    expect(chip.querySelector(".level")?.textContent).toBe("3");
    expect(chip.getAttribute("aria-label")).toBe("The Ring, tempted 3 times");
  });

  it("follows the level up", () => {
    const view = mountSeat([ring(1)]);
    expect(chipOf(view.container)!.querySelector(".level")?.textContent).toBe("1");
    expect(chipOf(view.container)!.getAttribute("aria-label")).toBe("The Ring, tempted once");
    view.setProps({ seat: seat([ring(2)]) } as never);
    expect(chipOf(view.container)!.querySelector(".level")?.textContent).toBe("2");
  });

  it("describes every line, gained or not, to a screen reader without opening", () => {
    const { container } = mountSeat([ring(3)]);
    const chip = chipOf(container)!;
    const desc = document.getElementById(chip.getAttribute("aria-describedby")!);
    expect(desc).not.toBeNull();
    expect(desc!.textContent).toContain("Gained: Your Ring-bearer is legendary");
    expect(desc!.textContent).toContain(
      "Not yet, after the 4th temptation: Whenever your Ring-bearer deals combat damage",
    );
    expect(cardOf()).toBeNull();
  });

  it("opens on a press with all four lines, the ones not yet gained dimmed and marked in words", () => {
    const { container } = mountSeat([ring(2)]);
    const chip = chipOf(container)!;
    expect(chip.getAttribute("aria-expanded")).toBe("false");
    chip.click();
    flushSync();
    expect(chip.getAttribute("aria-expanded")).toBe("true");

    const card = cardOf()!;
    expect(card).not.toBeNull();
    const items = [...card.querySelectorAll("li.line")];
    expect(items).toHaveLength(4);

    const gained = items.filter((li) => li.classList.contains("gained"));
    const pending = items.filter((li) => li.classList.contains("pending"));
    expect(gained.map((li) => li.querySelector(".text")?.textContent)).toEqual([
      LINES[0].text,
      LINES[1].text,
    ]);
    expect(pending.map((li) => li.querySelector(".when")?.textContent)).toEqual([
      "after the 3rd temptation",
      "after the 4th temptation",
    ]);
    // More than colour: a gained line has no "after the …" note and a
    // check mark; a pending one has the note and its count.
    for (const li of gained) {
      expect(li.querySelector(".when")).toBeNull();
      expect(li.querySelector(".mark svg")).not.toBeNull();
    }
    expect(pending.map((li) => li.querySelector(".mark")?.textContent?.trim())).toEqual(["3", "4"]);

    chip.click();
    flushSync();
    expect(cardOf()).toBeNull();
  });

  it("closes on Escape", () => {
    const { container } = mountSeat([ring(4)]);
    const chip = chipOf(container)!;
    chip.click();
    flushSync();
    expect(cardOf()).not.toBeNull();
    expect(cardOf()!.querySelectorAll("li.pending")).toHaveLength(0);
    chip.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(cardOf()).toBeNull();
  });

  it("keeps the plain chip for an emblem with no level", () => {
    const { container } = mountSeat([
      ring(1),
      { instance_id: "e2", label: "Elspeth emblem", text: "Creatures get +2/+2." },
    ]);
    expect(container.querySelectorAll(".emblem")).toHaveLength(2);
    expect(container.querySelectorAll("button.emblem.levelled")).toHaveLength(1);
    expect(container.querySelector("span.emblem")?.getAttribute("title")).toBe(
      "Elspeth emblem — Creatures get +2/+2.",
    );
  });
});

const creature = (extra: Partial<CardView>): CardView =>
  ({
    instance_id: "nazgul",
    name: "Nazgûl",
    owner: "alice",
    controller: "alice",
    known_by_you: true,
    type_line: "Creature — Zombie Wraith Knight",
    power: 1,
    toughness: 2,
    ...extra,
  }) as unknown as CardView;

function mountCard(card: CardView, ringBearerOf?: string) {
  return render(Card as never, { card, ringBearerOf } as Record<string, unknown>);
}

describe("the Ring-bearer marker", () => {
  it("is absent on a creature that is not a Ring-bearer", () => {
    const { container } = mountCard(creature({}));
    expect(container.querySelector(".ring-marker")).toBeNull();
  });

  it("marks the Ring-bearer, titled with its controller", () => {
    const { container } = mountCard(creature({ ring_bearer: true }), "Alice");
    const marker = container.querySelector(".ring-marker")!;
    expect(marker).not.toBeNull();
    expect(marker.getAttribute("aria-label")).toBe("Ring-bearer");
    expect(marker.getAttribute("title")).toBe("Alice's Ring-bearer");
    // The card's own name carries it too: the marker sits inside the
    // card's role, which a screen reader reads as one thing.
    expect(container.querySelector(".card")?.getAttribute("aria-label")).toBe(
      "Nazgûl, Alice's Ring-bearer",
    );
    // The PR 2 placeholder is gone.
    expect(container.querySelector(".badge.ring-bearer")).toBeNull();
  });

  it("shows beside MONSTROUS in the designation slot, not instead of it", () => {
    const { container } = mountCard(creature({ ring_bearer: true, monstrous: true }), "Alice");
    expect(container.querySelector(".ring-marker")).not.toBeNull();
    expect(container.querySelector(".badge.designation")?.textContent?.trim()).toBe("MONSTROUS");
  });

  it("shows beside HARNESSED, and beside CMD on a commander", () => {
    const { container } = mountCard(
      // A scryfall_id draws the art branch, which is where CMD lives.
      creature({ ring_bearer: true, harnessed: true, is_commander: true, scryfall_id: "s1" }),
      "Alice",
    );
    expect(container.querySelector(".ring-marker")).not.toBeNull();
    expect(container.querySelector(".badge.designation")?.textContent?.trim()).toBe("HARNESSED");
    expect(container.querySelector(".badge.cmd")).not.toBeNull();
  });

  it("shows on a face-down Ring-bearer (ADR 0114 §9)", () => {
    const { container } = mountCard(
      creature({ ring_bearer: true, face_down: true, known_by_you: false, name: "" }),
      "Alice",
    );
    expect(container.querySelector(".ring-marker")).not.toBeNull();
    expect(container.querySelector(".card")?.getAttribute("aria-label")).toBe(
      "face-down card, Alice's Ring-bearer",
    );
  });
});
