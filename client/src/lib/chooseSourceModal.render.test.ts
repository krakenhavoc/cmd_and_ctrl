// @vitest-environment jsdom
//
// ADR 0107 §6 (#1860): the "choose_source" prompt in ChoicePromptModal —
// "a source of your choice" (CR 609.7a), asked as a Circle of
// Protection's shield is made.
//
// It rides the bounded card-set grid with a floor and ceiling of one, so
// the bug a render test catches is the kind falling through to the
// fallback arm (which reads `count` and says "discard"). What is its own:
// each candidate is captioned with whose it is and which zone it is in,
// because the candidates span the battlefield, the stack and the
// graveyards.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { damageSourceCaption, damageSourceWhere } from "./damageSource";
import type { ActionType, CardView, GameView, PendingChoiceView } from "./protocol";
import { barButton, barPrimary, nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string, controller: string): CardView =>
  ({
    instance_id: id,
    name,
    controller,
    owner: controller,
    type_line: "Creature — Dragon",
    known_by_you: true,
  }) as CardView;

const dragon = card("c1", "Shivan Dragon", "them");
const bolt = card("c2", "Lightning Bolt", "them");
const mine = card("c3", "My Bear", "me");
const dead = card("c4", "Dead Goblin", "them");

const snapWith = (choice: Partial<PendingChoiceView>): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me", graveyard: { kind: "graveyard", count: 0, cards: [] } },
      { id: "them", name: "Bob", graveyard: { kind: "graveyard", count: 1, cards: [dead] } },
    ],
    battlefield: { kind: "battlefield", count: 2, cards: [dragon, mine] },
    stack: { kind: "stack", count: 1, cards: [bolt] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "choose_source",
        chooser: "me",
        from_player: "me",
        reason: "Circle of Protection: Red — choose a red source",
        choose_min: 1,
        choose_max: 1,
        options: [dragon, bolt, mine, dead],
        ...choice,
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(choice: Partial<PendingChoiceView> = {}): { container: HTMLElement; sent: Sent[] } {
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

describe("ChoicePromptModal — choose_source (ADR 0107 §6)", () => {
  it("renders every candidate and submits exactly one as card_ids", () => {
    const { container, sent } = mount();
    expect(cardButtons(container)).toHaveLength(4);
    const submit = barPrimary(container)!;
    expect(submit.disabled).toBe(true);
    expect(nameOf(submit)).toBe("Choose this source");

    click(cardButtons(container)[1]);
    // One of one: the rest lock out.
    expect(cardButtons(container)[0].disabled).toBe(true);
    expect(submit.disabled).toBe(false);
    click(submit);
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["c2"] });
  });

  it("names the rule and captions each source with whose and where", () => {
    const { container } = mount();
    expect(container.textContent).toContain("CR 609.7a");
    const captions = [...container.querySelectorAll(".source-caption")].map((e) =>
      e.textContent?.trim(),
    );
    expect(captions).toEqual([
      "Bob's · on the battlefield",
      "Bob's · on the stack",
      "Yours · on the battlefield",
      "Bob's · in a graveyard",
    ]);
  });

  it("offers no Clear button: the floor is one", () => {
    const { container } = mount();
    expect(barButton("Clear", container)).toBeNull();
  });
});

describe("damageSource", () => {
  const snap = snapWith({});
  it("finds the zone a candidate is in", () => {
    expect(damageSourceWhere(snap, "c1")).toBe("on the battlefield");
    expect(damageSourceWhere(snap, "c2")).toBe("on the stack");
    expect(damageSourceWhere(snap, "c4")).toBe("in a graveyard");
    expect(damageSourceWhere(snap, "nowhere")).toBe("");
  });
  it("says Yours for the viewer's own sources, and the seat's name to anyone else", () => {
    expect(damageSourceCaption(snap, mine, "me")).toBe("Yours · on the battlefield");
    expect(damageSourceCaption(snap, mine, null)).toBe("Me's · on the battlefield");
  });
});
