// @vitest-environment jsdom
//
// ADR 0189 decision 3 (#3538): Kozilek, the Broken Reality and Scroll of
// Fate ask a player which cards from their own hand to manifest. The
// prompt is an ordinary choose_cards that declares
// `choose_destination: "manifest"`, and the client says so: the
// confirm verb is "Manifest" and the sheet says what the pick does.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, GameView, PendingChoiceView } from "./protocol";
import { barPrimary } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string) => ({
  instance_id: id,
  name,
  type_line: "Creature — Bear",
  known_by_you: true,
});

const snapWith = (choice: Partial<PendingChoiceView>): GameView =>
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
        id: "choice-1",
        kind: "choose_cards",
        chooser: "me",
        from_player: "me",
        reason: "Choose two cards from your hand to manifest",
        options: [card("c1", "Alpha"), card("c2", "Beta"), card("c3", "Gamma")],
        choose_min: 2,
        choose_max: 2,
        ...choice,
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(choice: Partial<PendingChoiceView>): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(choice),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const cardButtons = (container: HTMLElement): HTMLButtonElement[] => [
  ...container.querySelectorAll<HTMLButtonElement>("button.card-pick"),
];

describe("ChoicePromptModal — the manifest pick (ADR 0189)", () => {
  it("confirms with the verb Manifest and sends card_ids", () => {
    const { container, sent } = mount({ choose_destination: "manifest" });
    const primary = barPrimary(container)!;
    expect(primary.textContent).toContain("Manifest");
    expect(primary.disabled).toBe(true);

    click(cardButtons(container)[0]);
    click(cardButtons(container)[2]);
    expect(primary.disabled).toBe(false);
    click(primary);

    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["c1", "c3"] });
  });

  it("says what the pick does", () => {
    const { container } = mount({ choose_destination: "manifest" });
    expect(container.textContent).toContain("face down as 2/2 creatures");
  });

  it("keeps the plain verb for a choose_cards with no destination", () => {
    const { container } = mount({});
    expect(barPrimary(container)!.textContent).not.toContain("Manifest");
    expect(container.textContent).not.toContain("face down as 2/2 creatures");
  });

  it("confirms an edict's pick with the verb Sacrifice (#3664)", () => {
    const { container, sent } = mount({
      choose_destination: "sacrifice",
      reason: "Choose a creature to sacrifice",
      choose_min: 1,
      choose_max: 1,
    });
    const primary = barPrimary(container)!;
    expect(primary.textContent).toContain("Sacrifice");
    expect(container.textContent).toContain("sacrificed once every player has chosen");

    click(cardButtons(container)[1]);
    click(primary);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["c2"] });
  });
});
