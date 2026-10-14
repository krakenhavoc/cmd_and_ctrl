// @vitest-environment jsdom
//
// companionSlot.render.test.ts — ADR 0172 §8. A seat's revealed
// companion sits beside its commanders, face up for everyone. The
// owner's carries the {3} special action as a button built from the
// server's own row: it sends `special_action { kind: "companion" }`
// when the row is available and is greyed when it is not.

import { describe, it, expect, afterEach } from "vitest";

import CompanionSlot from "./components/board/CompanionSlot.svelte";
import type { CardView } from "./protocol";
import type { MenuAction } from "./contextMenu.logic";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(() => cleanup());

const ME = "me";

function jegantha(available: boolean | undefined): CardView {
  return {
    instance_id: "jeg",
    name: "Jegantha, the Wellspring",
    owner: ME,
    controller: ME,
    is_companion: true,
    type_line: "Legendary Creature — Elemental Elk",
    mana_cost: "{4}{R/G}",
    special_actions:
      available === undefined
        ? undefined
        : [{ kind: "companion", label: "Put into your hand (companion)", cost: "{3}", available }],
  } as CardView;
}

describe("CompanionSlot", () => {
  it("shows another seat's companion with no button", () => {
    const { container } = render(
      CompanionSlot as never,
      {
        seat: { id: "them", name: "Them", companion: jegantha(undefined) },
      } as never,
    );
    const group = container.querySelector(
      "[role='group'][aria-label='Them companion, Jegantha, the Wellspring']",
    );
    expect(group).not.toBeNull();
    expect(group?.querySelector(".caption")?.textContent).toBe("Companion");
    expect(container.querySelector("button.take")).toBeNull();
  });

  it("sends the companion special action from the owner's button", () => {
    const sent: MenuAction[] = [];
    const { container } = render(
      CompanionSlot as never,
      {
        seat: { id: ME, name: "Me", companion: jegantha(true) },
        onSpecialAction: (a: MenuAction) => sent.push(a),
      } as never,
    );
    const button = container.querySelector<HTMLButtonElement>(
      "button[aria-label='Pay {3}: put Jegantha, the Wellspring into your hand (companion)']",
    );
    expect(button).not.toBeNull();
    expect(button!.disabled).toBe(false);
    click(button!);
    expect(sent).toHaveLength(1);
    expect(sent[0]).toMatchObject({
      type: "special_action",
      player: ME,
      params: { card_id: "jeg", kind: "companion", strict: true, auto_tap: true },
    });
  });

  it("greys the button outside the window and sends nothing", () => {
    const sent: MenuAction[] = [];
    const { container } = render(
      CompanionSlot as never,
      {
        seat: { id: ME, name: "Me", companion: jegantha(false) },
        onSpecialAction: (a: MenuAction) => sent.push(a),
      } as never,
    );
    const button = container.querySelector<HTMLButtonElement>("button.take");
    expect(button?.disabled).toBe(true);
    click(button!);
    expect(sent).toHaveLength(0);
  });

  it("renders nothing once the companion has been taken", () => {
    const { container } = render(
      CompanionSlot as never,
      {
        seat: { id: ME, name: "Me" },
      } as never,
    );
    expect(container.querySelector(".companion-slot")).toBeNull();
  });
});
