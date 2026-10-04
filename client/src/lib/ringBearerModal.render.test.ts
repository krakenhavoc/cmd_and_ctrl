// @vitest-environment jsdom
//
// ADR 0114 §4 (#2076): the "ring_bearer" prompt in ChoicePromptModal —
// "the Ring tempts you: choose a creature you control" (CR 701.54a).
//
// It rides the bounded card-set grid with a floor and ceiling of one, so
// the bug a render test catches is the kind falling through to the
// fallback arm (which reads `count` and says "discard"). The table's own
// Ring display is ADR 0114 PR 4; this pins only that a human can answer.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView, PendingChoiceView } from "./protocol";
import { barButton, barPrimary, nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const creature = (id: string, name: string): CardView =>
  ({
    instance_id: id,
    name,
    controller: "me",
    owner: "me",
    type_line: "Creature — Zombie Wraith Knight",
    known_by_you: true,
  }) as CardView;

const nazgul = creature("c1", "Nazgûl");
const wurm = creature("c2", "Craw Wurm");

const snapWith = (choice: Partial<PendingChoiceView>): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [{ id: "me", name: "Me", graveyard: { kind: "graveyard", count: 0, cards: [] } }],
    battlefield: { kind: "battlefield", count: 2, cards: [nazgul, wurm] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "ring_bearer",
        chooser: "me",
        from_player: "me",
        count: 1,
        reason: "choose your Ring-bearer",
        choose_min: 1,
        choose_max: 1,
        options: [nazgul, wurm],
        ...choice,
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWith({}),
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

describe("ChoicePromptModal — ring_bearer (ADR 0114 §4)", () => {
  it("offers each creature and submits exactly one as card_ids", () => {
    const { container, sent } = mount();
    expect(cardButtons(container)).toHaveLength(2);
    expect(container.textContent).toContain("CR 701.54");
    const submit = barPrimary(container)!;
    expect(submit.disabled).toBe(true);
    expect(nameOf(submit)).toBe("Choose");

    click(cardButtons(container)[1]);
    expect(cardButtons(container)[0].disabled).toBe(true);
    click(submit);
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["c2"] });
  });

  it("offers no Clear button: the floor is one", () => {
    const { container } = mount();
    expect(barButton("Clear", container)).toBeNull();
  });
});
