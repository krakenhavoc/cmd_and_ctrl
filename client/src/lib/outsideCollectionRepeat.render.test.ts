// @vitest-environment jsdom
//
// #3955 — an "any number" pick from the card database (Spawnsire of
// Ulamog's free cast) takes the same card name as often as it is clicked:
// a player may own two copies of an Eldrazi. A row's "-" takes one back.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import {
  allowsRepeatedNames,
  pickedCount,
  toggleCollectionPick,
  unpickOneCollection,
} from "./outsideCollection";
import { freeCastRequest, outsideCastRequest } from "./freeCastRequest";
import { get } from "svelte/store";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  outsideCastRequest.set(false);
  freeCastRequest.set(null);
});

const snap = (max: number, destination: string): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Them" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-spawnsire",
        kind: "outside_collection",
        chooser: "me",
        from_player: "me",
        reason: "Spawnsire of Ulamog — choose any number of Eldrazi cards",
        choose_min: 0,
        choose_max: max,
        choose_destination: destination,
        collection_options: [
          {
            instance_id: "",
            name: "Ulamog's Crusher",
            type_line: "Creature — Eldrazi",
            mana_cost: "{8}",
          },
          {
            instance_id: "",
            name: "Eldrazi Drone",
            type_line: "Creature — Eldrazi",
            mana_cost: "{3}",
          },
        ],
        options: [],
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(
  max: number,
  destination = "cast_free_outside",
): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snap(max, destination),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const rows = (c: HTMLElement): HTMLElement[] =>
  Array.from(c.querySelectorAll("button.collection-row"));
const buttonNamed = (c: HTMLElement, text: string): HTMLButtonElement | undefined =>
  Array.from(c.querySelectorAll("button")).find((b) => b.textContent?.trim().startsWith(text)) as
    | HTMLButtonElement
    | undefined;

describe("picking the same card more than once", () => {
  it("is allowed for an any-number pick and not for a pair", () => {
    expect(allowsRepeatedNames(100)).toBe(true);
    expect(allowsRepeatedNames(2)).toBe(false);
    expect(allowsRepeatedNames(1)).toBe(false);
  });

  it("adds a copy per click up to the ceiling, and takes one back", () => {
    let picked = toggleCollectionPick([], "Crusher", 100);
    picked = toggleCollectionPick(picked, "Crusher", 100);
    picked = toggleCollectionPick(picked, "Drone", 100);
    expect(picked).toEqual(["Crusher", "Crusher", "Drone"]);
    expect(pickedCount(picked, "Crusher")).toBe(2);
    expect(unpickOneCollection(picked, "Crusher")).toEqual(["Crusher", "Drone"]);
    expect(unpickOneCollection(picked, "Missing")).toEqual(picked);
    expect(toggleCollectionPick(["a", "b", "c"], "d", 3)).toEqual(["a", "b", "c"]);
  });

  it("keeps the toggle for a pair, where a name cannot repeat", () => {
    expect(toggleCollectionPick(["A", "B"], "A", 2)).toEqual(["B"]);
  });

  it("sends two copies of one card from the picker, and asks the board to open the cast", () => {
    const { container, sent } = mount(100);
    click(rows(container)[0]);
    click(rows(container)[0]);
    click(rows(container)[1]);
    expect(rows(container)[0].textContent).toContain("×2");
    click(
      buttonNamed(
        container,
        "Take Ulamog's Crusher and Ulamog's Crusher and Eldrazi Drone",
      ) as HTMLElement,
    );
    expect(sent).toEqual([
      {
        type: "resolve_choice",
        params: {
          choice_id: "choice-spawnsire",
          card_names: ["Ulamog's Crusher", "Ulamog's Crusher", "Eldrazi Drone"],
        },
      },
    ]);
    expect(get(outsideCastRequest)).toBe(true);
  });

  it("takes one copy back with the row's minus button", () => {
    const { container, sent } = mount(100);
    click(rows(container)[0]);
    click(rows(container)[0]);
    click(container.querySelector("button.collection-less") as HTMLElement);
    click(buttonNamed(container, "Take Ulamog's Crusher") as HTMLElement);
    expect(sent).toEqual([
      {
        type: "resolve_choice",
        params: { choice_id: "choice-spawnsire", card_names: ["Ulamog's Crusher"] },
      },
    ]);
  });

  it("a pick of one card still replaces, and shows no minus", () => {
    const { container } = mount(1, "outside");
    click(rows(container)[0]);
    click(rows(container)[1]);
    expect(container.querySelector("button.collection-less")).toBeNull();
    expect(buttonNamed(container, "Take Eldrazi Drone")).toBeDefined();
    expect(get(outsideCastRequest)).toBe(false);
  });
});
