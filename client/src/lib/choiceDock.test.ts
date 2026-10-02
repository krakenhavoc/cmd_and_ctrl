// choiceDock.test.ts — ADR 0111 Delivery PR 5: which pending choices
// are answered inline in the dock, and what the dock request says.
// The rendering and the answers are pinned through the Game route in
// choiceDock.render.test.ts; this is the pure half.

import { describe, it, expect, vi } from "vitest";

import {
  choiceRequest,
  gameOverRequest,
  INLINE_LABEL_MAX,
  INLINE_OPTION_MAX,
  isInlineChoice,
  voteRequest,
} from "./choiceDock";
import { dockKeyAction } from "./dock";
import type { PendingChoiceView } from "./protocol";

const choice = (over: Partial<PendingChoiceView>): PendingChoiceView =>
  ({
    id: "c1",
    kind: "trigger_prompt",
    chooser: "me",
    from_player: "me",
    count: 0,
    ...over,
  }) as PendingChoiceView;

const handlers = () => ({
  onAnswer: vi.fn(),
  onCoin: vi.fn(),
  onOption: vi.fn(),
  onLoop: vi.fn(),
});

describe("isInlineChoice", () => {
  it("takes the yes/no family, coin, loop, mana and colour", () => {
    for (const kind of [
      "trigger_prompt",
      "optional_replacement",
      "confirm",
      "may_cast",
      "entry_pay_life",
      "pay_unless",
      "coin_call",
      "loop_shortcut",
      "mana_pick",
      "choose_color",
    ]) {
      expect(isInlineChoice(choice({ kind })), kind).toBe(true);
    }
  });

  it("leaves every sheet kind to the modal", () => {
    for (const kind of [
      "scry",
      "surveil",
      "look_at_top",
      "put_in_library",
      "mode_pick",
      "trigger_order",
      "replacement_order",
      "damage_assignment",
      "choose_creature_type",
      "choose_card_name",
      "search_library",
      "sacrifice_choice",
      "choose_cards",
      "discard_from_hand",
    ]) {
      expect(isInlineChoice(choice({ kind })), kind).toBe(false);
    }
    expect(isInlineChoice(null)).toBe(false);
  });

  it("pay_unless with card picks or a tap list is a sheet", () => {
    expect(
      isInlineChoice(
        choice({ kind: "pay_unless", pay_cards: { action: "discard", count: 1, options: [] } }),
      ),
    ).toBe(false);
    expect(isInlineChoice(choice({ kind: "pay_unless", tap_cost: { key: "waterbend" } }))).toBe(
      false,
    );
  });

  it("an option pick is inline only when short: six options, no cards, short labels", () => {
    const opts = (n: number, label = "Draw a card") => Array.from({ length: n }, () => ({ label }));
    for (const kind of ["option_pick", "entry_controller"]) {
      expect(isInlineChoice(choice({ kind, pick_options: opts(INLINE_OPTION_MAX) }))).toBe(true);
      expect(isInlineChoice(choice({ kind, pick_options: opts(INLINE_OPTION_MAX + 1) }))).toBe(
        false,
      );
      expect(isInlineChoice(choice({ kind, pick_options: [] }))).toBe(false);
      expect(
        isInlineChoice(choice({ kind, pick_options: opts(2, "x".repeat(INLINE_LABEL_MAX)) })),
      ).toBe(true);
      expect(
        isInlineChoice(choice({ kind, pick_options: opts(2, "x".repeat(INLINE_LABEL_MAX + 1)) })),
      ).toBe(false);
      expect(
        isInlineChoice(
          choice({
            kind,
            pick_options: [{ label: "Pile 1", cards: [{ instance_id: "a", name: "A" }] }],
          } as Partial<PendingChoiceView>),
        ),
      ).toBe(false);
    }
  });
});

describe("choiceRequest", () => {
  it("is a choice, named by the reason, focusing its dialog", () => {
    const r = choiceRequest(
      choice({ reason: "Mulldrifter — draw two cards?" }),
      { sourceName: "x" },
      handlers(),
    );
    expect(r.rank).toBe("choice");
    expect(r.label).toBe("Mulldrifter — draw two cards?");
    expect(r.question).toBe("Mulldrifter — draw two cards?");
    expect(r.focus).toBe("dialog");
  });

  it("falls back to the modal's headings", () => {
    const ctx = { sourceName: "Smothering Tithe" };
    expect(choiceRequest(choice({}), ctx, handlers()).label).toBe("Smothering Tithe triggered");
    expect(
      choiceRequest(choice({ kind: "pay_unless", pay_cost: "{2}" }), ctx, handlers()).label,
    ).toBe("Smothering Tithe — pay {2}?");
    expect(choiceRequest(choice({ kind: "coin_call" }), ctx, handlers()).label).toBe(
      "Call the flip",
    );
  });

  it("never presses on Enter or Escape", () => {
    const kinds: Partial<PendingChoiceView>[] = [
      {},
      { kind: "optional_replacement" },
      { kind: "confirm" },
      { kind: "may_cast" },
      { kind: "entry_pay_life", pay_cost: "2 life" },
      { kind: "pay_unless", pay_cost: "{2}" },
      { kind: "coin_call", allow_stop: true },
      { kind: "loop_shortcut" },
      { kind: "mana_pick", color_options: ["G"] },
      { kind: "choose_color", color_options: ["G"] },
      { kind: "option_pick", pick_options: [{ label: "A" }] },
    ];
    for (const over of kinds) {
      const r = choiceRequest(
        choice(over),
        { sourceName: "x", loopIterations: 10, loopAnswerable: true },
        handlers(),
      );
      expect(dockKeyAction(r, "Enter"), over.kind).toBeNull();
      expect(dockKeyAction(r, "Escape"), over.kind).toBeNull();
    }
  });

  it("names Y and N on the yes/no buttons, and answers through onAnswer", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({ kind: "pay_unless", pay_cost: "{2}" }),
      { sourceName: "x" },
      h,
    );
    expect(r.primary?.label).toBe("Pay {2}");
    expect(r.primary?.keyShortcuts).toBe("Y");
    expect(r.secondary?.map((a) => [a.label, a.keyShortcuts])).toEqual([["Don't pay", "N"]]);
    r.primary!.onPress();
    r.secondary![0].onPress();
    expect(h.onAnswer.mock.calls).toEqual([[true], [false]]);
  });

  it("carries a refusal as Not accepted", () => {
    const r = choiceRequest(choice({}), { sourceName: "x", rejection: "nope" }, handlers());
    expect(r.refusal).toMatchObject({ tag: "Not accepted", text: "nope", tone: "danger" });
  });
});

describe("voteRequest and gameOverRequest", () => {
  it("a vote is a step: it keeps next, with the tally on each option", () => {
    const r = voteRequest({
      vote: {
        id: "v",
        topic: "Monarchy?",
        options: ["yes", "no"],
        initiator: "b",
        ballots: { a: 1, b: 1 },
      },
      viewerID: "a",
      seats: [
        { id: "a", name: "Ann" },
        { id: "b", name: "Bob" },
      ],
      onCast: () => {},
      onEnd: () => {},
    });
    expect(r.rank).toBe("step");
    expect(r.label).toBe("open vote");
    expect(r.detail).toBe("called by Bob");
    expect(r.row?.map((a) => [a.label, a.note, a.pressed])).toEqual([
      ["yes", "0", false],
      ["no", "2", true],
      ["end vote", undefined, undefined],
    ]);
  });

  it("game over is Back to lobby, with no Enter", () => {
    const back = vi.fn();
    const r = gameOverRequest(back);
    expect(r.rank).toBe("gameOver");
    expect(r.primary?.label).toBe("Back to lobby");
    expect(dockKeyAction(r, "Enter")).toBeNull();
    r.primary!.onPress();
    expect(back).toHaveBeenCalledOnce();
  });
});
