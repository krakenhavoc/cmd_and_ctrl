// @vitest-environment jsdom
// ADR 0175: later seats in a simultaneous permanent choice see earlier picks.

import { afterEach, describe, expect, it } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const permanent = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name,
    type_line: "Creature — Human",
    owner: "earlier",
    controller: "earlier",
    ...extra,
  }) as CardView;

function mount() {
  const snap = {
    id: "g",
    state: "active",
    seats: [
      { id: "earlier", name: "Alice" },
      { id: "empty", name: "Cara" },
      { id: "me", name: "Bob" },
    ],
    battlefield: {
      kind: "battlefield",
      count: 2,
      cards: [
        permanent("picked-public", "Grizzly Bears"),
        permanent("picked-down", "Hidden Face Name", {
          face_down: true,
          known_by_you: false,
          face_down_kind: "morphed",
        }),
      ],
    },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: -1, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-2",
        kind: "choose_cards",
        chooser: "me",
        from_player: "me",
        reason: "Liliana, Dreadhorde General — choose a creature to sacrifice",
        choose_min: 1,
        choose_max: 1,
        options: [
          permanent("current-a", "Current choice A"),
          permanent("current-b", "Current choice B"),
        ],
        earlier_choices: [
          { player_id: "earlier", card_ids: ["picked-public", "picked-down", "missing"] },
          { player_id: "empty", card_ids: [] },
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

describe("ChoicePromptModal — earlier simultaneous choices (ADR 0175)", () => {
  it("shows earlier picks beside the current chooser's unchanged selection", () => {
    const { container } = mount();
    expect(container.textContent).toContain("Earlier choices");
    expect(container.textContent).toContain("Alice");
    expect(container.textContent).toContain("Grizzly Bears");
    expect(container.textContent).toContain("Face-down permanent");
    expect(container.textContent).toContain("Selected permanent");
    expect(container.textContent).toContain("Chose nothing");
    expect(container.textContent).not.toContain("Hidden Face Name");

    const currentPick = [...container.querySelectorAll("button")].find((button) =>
      button.getAttribute("aria-label")?.includes("Current choice A"),
    );
    expect(currentPick).toBeDefined();
    click(currentPick!);
    expect(currentPick!.getAttribute("aria-pressed")).toBe("true");
    expect(container.textContent).toContain("1 / 1 selected");
  });
});
