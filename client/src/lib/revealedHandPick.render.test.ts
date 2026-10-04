// @vitest-environment jsdom
//
// ADR 0116 — the revealed-hand pick's card filter. "Target player
// reveals their hand. You choose a nonland card from it." The prompt
// shows the whole revealed hand, but only the cards listed in
// `eligible` can be picked; the hint names what may be chosen.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string, typeLine: string): CardView =>
  ({ instance_id: id, name, type_line: typeLine, known_by_you: true }) as unknown as CardView;

const forest = card("f", "Forest", "Basic Land — Forest");
const bolt = card("b", "Lightning Bolt", "Instant");

const snapWith = (eligible?: string[], label?: string): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Ana" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "precombat_main", step: "main1" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "discard_from_hand",
        chooser: "me",
        from_player: "them",
        count: 1,
        reason: "Thoughtseize",
        options: [forest, bolt],
        ...(eligible ? { eligible } : {}),
        ...(label ? { eligible_label: label } : {}),
      },
    ],
  }) as unknown as GameView;

function mount(eligible?: string[], label?: string) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(eligible, label),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const pickButton = (container: HTMLElement, name: string): HTMLButtonElement | undefined =>
  [...container.querySelectorAll<HTMLButtonElement>(".card-pick")].find(
    (b) => b.getAttribute("aria-label") === `select ${name}`,
  );

const buttonNamed = (container: HTMLElement, text: string): HTMLButtonElement | undefined =>
  [...container.querySelectorAll("button")].find((b) => nameOf(b) === text);

describe("ChoicePromptModal — discard_from_hand with a filter (ADR 0116)", () => {
  it("shows the whole hand but greys out the cards that can't be chosen", () => {
    const { container, sent } = mount(["b"], "nonland card");

    const land = pickButton(container, "Forest");
    const spell = pickButton(container, "Lightning Bolt");
    expect(land).toBeDefined();
    expect(spell).toBeDefined();
    expect(land!.disabled).toBe(true);
    expect(spell!.disabled).toBe(false);

    click(land!);
    expect(land!.getAttribute("aria-pressed")).toBe("false");

    click(spell!);
    click(buttonNamed(container, "Confirm")!);
    expect(sent).toHaveLength(1);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["b"] });
  });

  it("names what may be chosen", () => {
    const { container } = mount(["b"], "nonland card");
    expect(container.textContent).toContain("Pick a nonland card from");
  });

  it("offers every card when the prompt carries no filter", () => {
    const { container } = mount();
    expect(pickButton(container, "Forest")!.disabled).toBe(false);
    expect(container.textContent).toContain("Pick 1 card from");
  });
});
