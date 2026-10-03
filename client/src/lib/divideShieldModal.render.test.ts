// @vitest-environment jsdom
//
// ADR 0108 §7 (#1904): the "divide_shield" prompt in ChoicePromptModal —
// CR 615.7's "the player … chooses which damage the shield prevents",
// asked when a charged shield ("prevent the next 3 damage") meets more
// damage at once than it can cover.
//
// The bug a render test catches is the kind falling through to the
// fallback arm, which would leave the table waiting on a prompt nobody
// can answer. What is its own: one number per damage event, the submit
// held until the numbers add up to the charge, and the answer sent as a
// distribution keyed by entry id.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, GameView, PendingChoiceView } from "./protocol";
import { barPrimary, nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const snap = (choice: Partial<PendingChoiceView> = {}): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Bob" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: {
      number: 3,
      active_seat: 1,
      priority_holder: -1,
      phase: "combat",
      step: "combat_damage",
    },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "divide_shield",
        chooser: "me",
        from_player: "me",
        reason: "Mending Hands — divide 3 prevention among the damage",
        divide_shield: {
          label: "Mending Hands",
          charge: 3,
          entries: [
            {
              id: "e1",
              source_id: "a",
              source_name: "Grizzly Bears",
              target_id: "me",
              target_name: "Me",
              target_is_player: true,
              amount: 2,
              combat: true,
            },
            {
              id: "e2",
              source_id: "b",
              source_name: "Hill Giant",
              target_id: "me",
              target_name: "Me",
              target_is_player: true,
              amount: 3,
              combat: true,
            },
          ],
        },
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
      snap: snap(),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

function setShare(input: HTMLInputElement, n: number): void {
  input.value = String(n);
  input.dispatchEvent(new Event("input", { bubbles: true }));
}

const inputs = (c: HTMLElement): HTMLInputElement[] => [
  ...c.querySelectorAll<HTMLInputElement>(".assign-input input"),
];

describe("ChoicePromptModal — divide_shield (ADR 0108 §7)", () => {
  it("shows one row per damage event and names the rule", async () => {
    const { container } = mount();
    expect(inputs(container)).toHaveLength(2);
    expect(container.textContent).toContain("CR 615.7");
    expect(container.textContent).toContain("Grizzly Bears → you: 2 combat damage");
    expect(container.textContent).toContain("Hill Giant → you: 3 combat damage");
  });

  it("holds the answer until the shares add up to the charge, then sends a distribution", async () => {
    const { container, sent } = mount();
    const submit = barPrimary(container)!;
    expect(nameOf(submit)).toBe("Prevent");
    expect(submit.disabled).toBe(true);

    const [a, b] = inputs(container);
    setShare(a, 2);
    await Promise.resolve();
    expect(barPrimary(container)!.disabled).toBe(true);
    setShare(b, 1);
    await Promise.resolve();
    expect(barPrimary(container)!.disabled).toBe(false);

    click(barPrimary(container)!);
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", distribution: { e1: 2, e2: 1 } });
  });

  it("never lets a share exceed its event", async () => {
    const { container } = mount();
    const [a] = inputs(container);
    setShare(a, 9);
    await Promise.resolve();
    expect(container.textContent).toContain("2 / 3 prevented");
  });
});
