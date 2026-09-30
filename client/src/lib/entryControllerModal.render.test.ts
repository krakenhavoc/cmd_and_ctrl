// @vitest-environment jsdom
//
// ADR 0102 / CR 614.12a — "this enchantment enters under the control of
// an opponent of your choice". The prompt rides option_pick's seat
// buttons and its {option_index} answer; what it adds is the sentence,
// worded by control_purpose.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const snapWithEntryController = (purpose: string): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "b", name: "Bea" },
      { id: "c", name: "Cy" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "precombat_main", step: "main1" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "entry_controller",
        chooser: "me",
        from_player: "me",
        reason: "Captive Audience — choose an opponent to control it",
        control_purpose: purpose,
        pick_options: [
          { label: "Bea", player: "b" },
          { label: "Cy", player: "c" },
        ],
      },
    ],
  }) as unknown as GameView;

function mount(purpose: string) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const view = render(
    ChoicePromptModal as never,
    {
      snap: snapWithEntryController(purpose),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

describe("ChoicePromptModal — entry_controller (CR 614.12a)", () => {
  it("renders one button per opponent under the card's question", () => {
    const { container } = mount("harm");

    expect(container.querySelector("#choice-title")?.textContent).toContain("Captive Audience");
    const labels = [...container.querySelectorAll(".pick-option")].map((b) =>
      (b.textContent ?? "").trim(),
    );
    expect(labels).toEqual(["Bea", "Cy"]);
  });

  it("answers with the index of the seat clicked", () => {
    const { container, sent } = mount("harm");

    click(container.querySelectorAll<HTMLButtonElement>(".pick-option")[1]);

    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", option_index: 1 });
  });

  it("words a helpful permanent as a gift", () => {
    const { container } = mount("benefit");

    expect(container.querySelector(".prompt-hint")?.textContent).toContain("get what it does");
  });
});
