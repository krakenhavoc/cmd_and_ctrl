// choiceDock.turnSkipOrder.test.ts — ADR 0188 Decision 12 (#3562): the
// which-skip question, as a dock request. Both answers skip the extra
// turn; "Use my skip" (answer 0) is recommended, takes focus so Enter
// takes it, and is what a remembered Always means. The wire is
// pending_choices[].turn_skip_order (#3561, PR #3752).

import { describe, it, expect, vi } from "vitest";

import { choiceRequest, isInlineChoice } from "./choiceDock";
import { dockKeyAction } from "./dock";
import { TURN_SKIP_ORDER_KEY, canRemember, rememberedRule } from "./autoAnswerPref";
import { L } from "./labels";
import type { PendingChoiceView } from "./protocol";

const choice = (over: Partial<PendingChoiceView> = {}): PendingChoiceView =>
  ({
    id: "c1",
    kind: "trigger_prompt",
    chooser: "me",
    from_player: "me",
    count: 0,
    ...over,
  }) as PendingChoiceView;

const skip = (over: Partial<PendingChoiceView> = {}): PendingChoiceView =>
  choice({
    kind: "turn_skip_order",
    reason:
      'Your extra turn is skipped either way. Use it to pay off your own "skip your next turn"?',
    accept_label: "Use my skip (Magosi, the Waterveil)",
    decline_label: "Let Trouble in Pairs skip it",
    auto_answer_key: TURN_SKIP_ORDER_KEY,
    auto_answer_card: "Magosi, the Waterveil",
    auto_answer_prompt: "Use your own skip on an extra turn another effect would skip?",
    turn_skip_order: {
      extra_turn_ref: 1,
      skip_source_name: "Magosi, the Waterveil",
      others: ["Trouble in Pairs"],
      recommended: 0,
    },
    ...over,
  });

const handlers = () => ({
  onAnswer: vi.fn(),
  onCoin: vi.fn(),
  onOption: vi.fn(),
  onLoop: vi.fn(),
});

describe("turn_skip_order, the which-skip question (ADR 0188)", () => {
  it("is answered inline in the dock", () => {
    expect(isInlineChoice(skip())).toBe(true);
  });

  it("words the two answers from the server's labels, and tags answer 0 Recommended", () => {
    const r = choiceRequest(skip(), { sourceName: "x" }, handlers());
    expect(r.question).toContain("Your extra turn is skipped either way");
    expect([r.primary?.label, r.primary?.note]).toEqual([
      "Use my skip (Magosi, the Waterveil)",
      "Recommended",
    ]);
    expect(L.recommendedAnswer).toBe("Recommended");
    expect(r.secondary?.map((a) => [a.label, a.note])).toEqual([
      ["Let Trouble in Pairs skip it", undefined],
    ]);
  });

  it("opens with focus on answer 0, and binds no Enter or Escape of its own", () => {
    const r = choiceRequest(skip(), { sourceName: "x" }, handlers());
    expect(r.focus).toBe("primary");
    expect(r.primary?.id).toBe("yes");
    // The first button is focused rather than bound: only an Enter on
    // the button takes it, never one pressed elsewhere at the table.
    expect(dockKeyAction(r, "Enter")).toBeNull();
    expect(dockKeyAction(r, "Escape")).toBeNull();
    // Every other yes/no keeps focus on the dialog.
    expect(choiceRequest(choice(), { sourceName: "x" }, handlers()).focus).toBe("dialog");
  });

  it("sends apply true for answer 0 and false for answer 1", () => {
    const h = handlers();
    const r = choiceRequest(skip(), { sourceName: "x" }, h);
    expect([r.primary?.keyShortcuts, r.secondary?.[0].keyShortcuts]).toEqual(["U", "L"]);
    r.primary!.onPress();
    r.secondary![0].onPress();
    expect(h.onAnswer.mock.calls).toEqual([[true], [false]]);
  });

  it("offers Remember this answer, and Always is answer 0", () => {
    const onToggle = vi.fn();
    const r = choiceRequest(
      skip(),
      { sourceName: "x", remember: { on: false, onToggle } },
      handlers(),
    );
    expect(canRemember(skip())).toBe(true);
    const toggle = r.row?.find((a) => a.id === "remember");
    expect(toggle?.label).toBe(L.rememberThisAnswer);
    toggle?.onPress();
    expect(onToggle).toHaveBeenCalled();
    expect(rememberedRule(skip(), true)).toMatchObject({
      key: "engine|turn_skip_order",
      answer: "always",
    });
    expect(rememberedRule(skip(), false)?.answer).toBe("never");
  });

  it("falls back to plain words when the server sends no labels", () => {
    const r = choiceRequest(
      skip({ accept_label: undefined, decline_label: undefined, reason: "" }),
      { sourceName: "x" },
      handlers(),
    );
    expect(r.primary?.label).toBe("Use my skip");
    expect(r.secondary?.[0].label).toBe("Let the other effect skip it");
    expect(r.question).toMatch(/skipped either way/);
  });
});
