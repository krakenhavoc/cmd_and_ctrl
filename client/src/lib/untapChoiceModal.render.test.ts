// @vitest-environment jsdom
//
// #826 / CR 502.3 — the untap step's own determination in the picker.
//
// Two things about this prompt are decisions rather than plumbing, and
// both live in the markup, so they need a DOM:
//
//   - A prompt whose ceiling is "all of them" is a board of nothing but
//     "you may choose NOT to untap" permanents, and the default there is
//     to untap. It opens PRE-FILLED, so the player's click is the
//     deselection.
//   - A prompt where a cap binds has no such default and opens empty,
//     because pre-filling it would hand the player an illegal set to
//     undo.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string): CardView =>
  ({ instance_id: id, name, type_line: "Land", tapped: true }) as unknown as CardView;

const snapWithUntapChoice = (chooseMin: number, chooseMax: number, options: CardView[]): GameView =>
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
    turn: { number: 3, active_seat: 0, priority_holder: -1, phase: "beginning", step: "untap" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "untap_choice",
        chooser: "me",
        from_player: "me",
        reason: "Untap step — choose which permanents untap",
        choose_min: chooseMin,
        choose_max: chooseMax,
        options,
      },
    ],
  }) as unknown as GameView;

function mount(chooseMin: number, chooseMax: number, options: CardView[]) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const view = render(
    ChoicePromptModal as never,
    {
      snap: snapWithUntapChoice(chooseMin, chooseMax, options),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const selected = (container: HTMLElement): string[] =>
  [...container.querySelectorAll('button[aria-pressed="true"]')].map(
    (b) => b.getAttribute("aria-label") ?? "",
  );

const untapButton = (container: HTMLElement): HTMLButtonElement | undefined =>
  [...container.querySelectorAll("button")].find((b) => (b.textContent ?? "").trim() === "Untap");

describe("ChoicePromptModal — untap_choice (CR 502.3)", () => {
  it("renders the card picker with the step's own header", () => {
    const { container } = mount(1, 1, [card("a", "Forest"), card("b", "Island")]);

    const title = container.querySelector("#choice-title");
    expect(title?.textContent).toContain("Untap step");
    expect(title?.textContent).toContain("CR 502.3");
    expect(container.querySelectorAll(".card-pick")).toHaveLength(2);
  });

  it("opens empty when a cap binds, and submits the picked permanent", () => {
    const { container, sent } = mount(1, 1, [card("a", "Forest"), card("b", "Island")]);

    expect(selected(container)).toHaveLength(0);
    // Nothing chosen is not a legal answer under a cap, so the button
    // is not available until something is.
    expect(untapButton(container)?.disabled).toBe(true);

    click([...container.querySelectorAll(".card-pick")][1] as HTMLElement);
    expect(untapButton(container)?.disabled).toBe(false);
    click(untapButton(container)!);

    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["b"] });
  });

  it("opens pre-filled when everything may be left tapped, so the click is the opt-out", () => {
    const { container, sent } = mount(0, 2, [card("a", "Rust Tick"), card("b", "Amber Prison")]);

    expect(selected(container)).toHaveLength(2);

    // Deselect one: "you may choose not to untap this artifact".
    click([...container.querySelectorAll(".card-pick")][0] as HTMLElement);
    click(untapButton(container)!);

    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["b"] });
  });
});
