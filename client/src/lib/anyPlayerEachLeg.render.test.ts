// @vitest-environment jsdom
// ADR 0187 §8: one player's leg of a per-land any-player offer (Cleansing).

import { afterEach, describe, expect, it } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const land = (id: string, name: string, controller: string): CardView =>
  ({
    instance_id: id,
    name,
    type_line: "Basic Land — Forest",
    owner: controller,
    controller,
  }) as CardView;

function mount() {
  const lands = [
    land("a1", "Alice Forest", "alice"),
    land("m1", "My Island", "me"),
    land("m2", "My Swamp", "me"),
    land("m3", "My Plains", "me"),
  ];
  const snap = {
    id: "g",
    state: "active",
    seats: [
      { id: "alice", name: "Alice" },
      { id: "bob", name: "Bob" },
      { id: "me", name: "Carol" },
    ],
    battlefield: { kind: "battlefield", count: lands.length, cards: lands },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: -1, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-9",
        kind: "choose_cards",
        chooser: "me",
        from_player: "me",
        reason: "Cleansing — choose lands to pay 1 life for",
        choose_min: 0,
        choose_max: 3,
        options: lands,
        // The engine's suggestion: my own lands nobody chose yet.
        choose_suggested: ["m2", "m3"],
        any_player_offer: {
          controller: "bob",
          order: ["alice", "bob", "me"],
          declined: ["bob"],
          each_life: 1,
        },
        earlier_choices: [
          { player_id: "alice", card_ids: ["a1", "m1"] },
          { player_id: "bob", card_ids: [] },
        ],
      },
    ],
  } as unknown as GameView;
  const sent: { type: ActionType; params?: unknown }[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap,
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const pick = (container: HTMLElement, name: string) =>
  [...container.querySelectorAll("button")].find((b) =>
    b.getAttribute("aria-label")?.includes(name),
  )!;

describe("ChoicePromptModal — a per-land any-player offer (ADR 0187 §8)", () => {
  it("says whose effect it is, who is next and what the earlier players chose", () => {
    const { container } = mount();
    const text = container.textContent ?? "";
    expect(text).toContain("Bob's");
    expect(text).toContain("Each player in turn chooses lands to pay 1 life for");
    expect(text).toContain("You are the last to choose.");
    expect(text).toContain("So far: Alice pays for 2 lands, Bob for none.");
    expect(text).toContain("Nothing is paid until the last player has chosen.");
    // The faces section is for the other choose_cards shapes.
    expect(text).not.toContain("Earlier choices");
  });

  it("badges a land an earlier player pays for", () => {
    const { container } = mount();
    const badges = [...container.querySelectorAll(".pay-badge")].map((e) => e.textContent?.trim());
    expect(badges).toEqual(["Alice pays", "Alice pays"]);
    expect(pick(container, "Alice Forest").textContent).toContain("Alice pays");
    expect(pick(container, "My Swamp").textContent).not.toContain("pays");
  });

  it("opens on the suggested lands and prices them", () => {
    const { container } = mount();
    expect(pick(container, "My Swamp").getAttribute("aria-pressed")).toBe("true");
    expect(pick(container, "My Plains").getAttribute("aria-pressed")).toBe("true");
    expect(pick(container, "My Island").getAttribute("aria-pressed")).toBe("false");
    expect(container.querySelector(".each-price")?.textContent?.trim()).toBe(
      "Pay 2 life for 2 lands",
    );
    expect(container.textContent).toContain("2 / 3 selected");
  });

  it("warns on a selected land someone already pays for, and the price follows the toggles", () => {
    const { container } = mount();
    expect(container.querySelector(".pay-warn")).toBeNull();
    click(pick(container, "My Island"));
    const warn = container.querySelector(".pay-warn");
    expect(warn?.textContent?.trim()).toBe("Alice already pays for this: you'd both pay 1 life");
    expect(container.querySelector(".each-price")?.textContent?.trim()).toBe(
      "Pay 3 life for 3 lands (1 already paid for by others)",
    );
    click(pick(container, "My Island"));
    click(pick(container, "My Swamp"));
    click(pick(container, "My Plains"));
    expect(container.querySelector(".pay-warn")).toBeNull();
    expect(container.querySelector(".each-price")?.textContent?.trim()).toBe("Pay nothing");
  });

  it("answers with the chosen lands", () => {
    const { container, sent } = mount();
    click(pick(container, "My Plains"));
    const confirm = [...container.querySelectorAll("button")].find((b) =>
      b.textContent?.trim().startsWith("Pay 1 life"),
    );
    expect(confirm).toBeDefined();
    click(confirm!);
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-9", card_ids: ["m2"] });
  });
});
