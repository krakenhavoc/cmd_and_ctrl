// @vitest-environment jsdom
//
// #4947 — a pick from the card database "with different names" (Turtles
// Forever's outside half, up to four legendary creature cards) takes each
// name once, whatever its ceiling: a second click on a picked name takes
// it back instead of adding a copy the server would refuse. Research's
// "up to four cards", where a player may own two copies, still repeats.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { allowsRepeatedNames, toggleCollectionPick } from "./outsideCollection";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const snap = (distinct: boolean): GameView =>
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
        id: "choice-outside",
        kind: "outside_collection",
        chooser: "me",
        from_player: "me",
        reason: distinct
          ? "Turtles Forever — choose legendary creature cards you own from outside the game with different names"
          : "Research — choose up to four cards you own from outside the game",
        choose_min: 0,
        choose_max: 4,
        choose_destination: "outside",
        ...(distinct ? { distinct_names: true } : {}),
        collection_options: [
          {
            instance_id: "",
            name: "Leonardo, Leader in Blue",
            type_line: "Legendary Creature — Mutant Ninja Turtle",
            mana_cost: "{2}{W}",
          },
          {
            instance_id: "",
            name: "Raphael, Fiery Fighter",
            type_line: "Legendary Creature — Mutant Ninja Turtle",
            mana_cost: "{2}{R}",
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

function mount(distinct: boolean): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snap(distinct),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const rows = (c: HTMLElement): HTMLButtonElement[] =>
  Array.from(c.querySelectorAll("button.collection-row"));
const buttonNamed = (c: HTMLElement, text: string): HTMLButtonElement | undefined =>
  Array.from(c.querySelectorAll("button")).find((b) => b.textContent?.trim().startsWith(text)) as
    | HTMLButtonElement
    | undefined;

describe("a pick with different names", () => {
  it("never repeats a name, whatever the ceiling", () => {
    expect(allowsRepeatedNames(4, true)).toBe(false);
    expect(allowsRepeatedNames(100, true)).toBe(false);
    expect(allowsRepeatedNames(4)).toBe(true);
    expect(toggleCollectionPick(["A"], "A", 4, true)).toEqual([]);
    expect(toggleCollectionPick(["A"], "B", 4, true)).toEqual(["A", "B"]);
    expect(toggleCollectionPick(["A"], "A", 4)).toEqual(["A", "A"]);
  });

  it("takes a picked name back on a second click, with no copy count or minus", () => {
    const { container } = mount(true);
    click(rows(container)[0]);
    expect(rows(container)[0].getAttribute("aria-pressed")).toBe("true");
    click(rows(container)[0]);
    expect(rows(container)[0].getAttribute("aria-pressed")).toBe("false");
    expect(container.querySelector(".collection-copies")).toBeNull();
    expect(container.querySelector("button.collection-less")).toBeNull();
    expect(buttonNamed(container, "Take…")?.disabled).toBe(true);
  });

  it("sends each picked name once", () => {
    const { container, sent } = mount(true);
    click(rows(container)[0]);
    click(rows(container)[0]);
    click(rows(container)[0]);
    click(rows(container)[1]);
    expect(container.querySelector(".collection-copies")).toBeNull();
    click(
      buttonNamed(
        container,
        "Take Leonardo, Leader in Blue and Raphael, Fiery Fighter",
      ) as HTMLElement,
    );
    expect(sent).toEqual([
      {
        type: "resolve_choice",
        params: {
          choice_id: "choice-outside",
          card_names: ["Leonardo, Leader in Blue", "Raphael, Fiery Fighter"],
        },
      },
    ]);
  });

  it("leaves Research's pick, where a name may repeat, as it was", () => {
    const { container, sent } = mount(false);
    click(rows(container)[0]);
    click(rows(container)[0]);
    expect(rows(container)[0].textContent).toContain("×2");
    expect(container.querySelector("button.collection-less")).not.toBeNull();
    click(
      buttonNamed(
        container,
        "Take Leonardo, Leader in Blue and Leonardo, Leader in Blue",
      ) as HTMLElement,
    );
    expect(sent).toEqual([
      {
        type: "resolve_choice",
        params: {
          choice_id: "choice-outside",
          card_names: ["Leonardo, Leader in Blue", "Leonardo, Leader in Blue"],
        },
      },
    ]);
  });
});
