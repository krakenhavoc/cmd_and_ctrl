// @vitest-environment jsdom
//
// #1198 / CR 614.1c — "as this land enters, you may reveal an Island
// or Swamp card from your hand. If you don't, it enters tapped."
//
// The kind rides the same bounded card-set grid as choose_cards, so
// what needs a DOM is the two things that are decisions rather than
// plumbing:
//
//   - The picker opens EMPTY and the decline is always reachable. The
//     floor is zero by design — showing nothing is a real answer, and
//     one a player may want even holding a match (bluffing an empty
//     hand). The untap prompt pre-fills at a ceiling of "all of them";
//     this one must not, or the click that means "reveal" would be the
//     click that means "don't".
//   - Both answers leave through the ordinary {choice_id, card_ids}
//     payload, the empty one included.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const handCard = (id: string, name: string): CardView =>
  ({ instance_id: id, name, type_line: "Basic Land — Island" }) as unknown as CardView;

const snapWithEntryReveal = (options: CardView[]): GameView =>
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
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "precombat_main", step: "main1" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "entry_reveal_from_hand",
        chooser: "me",
        from_player: "me",
        reason:
          "Choked Estuary — reveal an Island or Swamp card from your hand so it enters untapped?",
        choose_min: 0,
        choose_max: 1,
        options,
      },
    ],
  }) as unknown as GameView;

function mount(options: CardView[]) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const view = render(
    ChoicePromptModal as never,
    {
      snap: snapWithEntryReveal(options),
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

const buttonNamed = (container: HTMLElement, text: string): HTMLButtonElement | undefined =>
  [...container.querySelectorAll("button")].find((b) => (b.textContent ?? "").trim() === text);

describe("ChoicePromptModal — entry_reveal_from_hand (CR 614.1c)", () => {
  it("renders the card picker with the land's own question", () => {
    const { container } = mount([handCard("a", "Island"), handCard("b", "Swamp")]);

    const title = container.querySelector("#choice-title");
    expect(title?.textContent).toContain("Choked Estuary");
    expect(container.querySelectorAll(".card-pick")).toHaveLength(2);
  });

  it("opens empty, so the first click reveals rather than un-reveals", () => {
    const { container } = mount([handCard("a", "Island"), handCard("b", "Swamp")]);

    expect(selected(container)).toHaveLength(0);
  });

  it("sends the decline as an empty card_ids list", () => {
    const { container, sent } = mount([handCard("a", "Island")]);

    const decline = buttonNamed(container, "Reveal nothing");
    expect(decline).toBeDefined();
    click(decline!);

    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: [] });
  });

  it("sends the revealed card when one is picked", () => {
    const { container, sent } = mount([handCard("a", "Island"), handCard("b", "Swamp")]);

    click(container.querySelectorAll<HTMLButtonElement>(".card-pick")[1]);
    const reveal = buttonNamed(container, "Reveal");
    expect(reveal).toBeDefined();
    click(reveal!);

    expect(sent).toHaveLength(1);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["b"] });
  });
});
