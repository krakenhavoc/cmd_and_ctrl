// @vitest-environment jsdom
//
// #2115 — the revealed-hand pick's variants: "You may choose a nonland
// card from it" (an empty answer is a real one), "… and exile that
// card" (exiling is not discarding), and Agonizing Remorse's "or a card
// from their graveyard".

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
const swamp = card("s", "Swamp", "Basic Land — Swamp");

interface Variant {
  destination: "discard" | "exile";
  optional: boolean;
  graveyard?: boolean;
}

const snapWith = (v: Variant): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me", graveyard: { kind: "graveyard", count: 0, cards: [] } },
      {
        id: "them",
        name: "Ana",
        graveyard: {
          kind: "graveyard",
          count: v.graveyard ? 1 : 0,
          cards: v.graveyard ? [swamp] : [],
        },
      },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "precombat_main", step: "main1" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "revealed_hand_pick",
        chooser: "me",
        from_player: "them",
        count: 1,
        reason: "Nightsnare",
        options: v.graveyard ? [forest, bolt, swamp] : [forest, bolt],
        eligible: v.graveyard ? ["b", "s"] : ["b"],
        eligible_label: "nonland card",
        ...(v.optional ? {} : { choose_min: 1 }),
        choose_max: 1,
        pick_destination: v.destination,
        ...(v.graveyard ? { pick_from_graveyard: true } : {}),
      },
    ],
  }) as unknown as GameView;

function mount(v: Variant) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(v),
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

describe("ChoicePromptModal — revealed_hand_pick (#2115)", () => {
  it("lets an optional pick choose nothing", () => {
    const { container, sent } = mount({ destination: "discard", optional: true });
    const none = buttonNamed(container, "Choose nothing");
    expect(none).toBeDefined();
    expect(none!.disabled).toBe(false);
    expect(container.textContent).toContain("You may choose nothing");
    click(none!);
    expect(sent).toHaveLength(1);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: [] });
  });

  it("confirms a chosen card, and greys out the rest of the hand", () => {
    const { container, sent } = mount({ destination: "discard", optional: true });
    expect(pickButton(container, "Forest")!.disabled).toBe(true);
    click(pickButton(container, "Lightning Bolt")!);
    expect(buttonNamed(container, "Choose nothing")).toBeUndefined();
    click(buttonNamed(container, "Confirm")!);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["b"] });
  });

  it("does not offer choosing nothing when the pick is mandatory", () => {
    const { container } = mount({ destination: "exile", optional: false });
    expect(buttonNamed(container, "Choose nothing")).toBeUndefined();
    expect(buttonNamed(container, "Confirm")!.disabled).toBe(true);
  });

  it("says an exiled pick is not discarded", () => {
    const { container } = mount({ destination: "exile", optional: false });
    expect(container.textContent).toContain("Your pick is exiled");
    expect(container.textContent).not.toContain("will discard your pick");
  });

  it("offers the graveyard after the hand, captioned", () => {
    const { container, sent } = mount({ destination: "exile", optional: false, graveyard: true });
    expect(container.textContent).toContain("or any card from their graveyard");
    const land = pickButton(container, "Swamp");
    expect(land).toBeDefined();
    expect(land!.disabled).toBe(false);
    expect(land!.textContent).toContain("in graveyard");
    click(land!);
    click(buttonNamed(container, "Confirm")!);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["s"] });
  });
});
