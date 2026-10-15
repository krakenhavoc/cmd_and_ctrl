import { describe, it, expect } from "vitest";

import { askedByHandText, autoAnswerNotice, autoAnswerNoticeRequest } from "./autoAnswerPref";
import { L } from "./labels";
import type { GameView, LogEvent, PendingChoiceView } from "./protocol";
import {
  canRememberTriggerOrder,
  isTriggerOrderKey,
  keepsRowsTogether,
  triggerOrderAnswer,
} from "./triggerOrderMemory";

// triggerOrderMemory.test.ts — "Use this order again" (#3016, ADR 0018's
// eighth amendment): the toggle's rules, the answer it sends, and the
// automatic-answer notice for a remembered order.

const opts = [
  { id: "p1", label: "Plunderer", row: 1 },
  { id: "n", label: "Noble", row: 2 },
  { id: "p2", label: "Plunderer", row: 1 },
];

const keyed = (over: Partial<PendingChoiceView> = {}): PendingChoiceView =>
  ({
    id: "c",
    kind: "trigger_order",
    chooser: "me",
    auto_answer_key: "trigger_order|noble+plunderer",
    trigger_options: opts,
    ...over,
  }) as PendingChoiceView;

describe("keepsRowsTogether", () => {
  it("takes an order that keeps each row's copies together", () => {
    expect(keepsRowsTogether(["p1", "p2", "n"], opts)).toBe(true);
    expect(keepsRowsTogether(["n", "p2", "p1"], opts)).toBe(true);
  });
  it("refuses an order that splits a row", () => {
    expect(keepsRowsTogether(["p1", "n", "p2"], opts)).toBe(false);
  });
  it("judges a partial order by what is placed so far", () => {
    expect(keepsRowsTogether(["p1"], opts)).toBe(true);
    expect(keepsRowsTogether(["p1", "n"], opts)).toBe(true);
  });
  it("never splits options with no row", () => {
    expect(keepsRowsTogether(["a", "b", "a"], [{ id: "a" }, { id: "b" }])).toBe(true);
  });
});

describe("triggerOrderAnswer", () => {
  it("sends only the order on a prompt with no key", () => {
    const c = keyed({ auto_answer_key: undefined });
    expect(canRememberTriggerOrder(c)).toBe(false);
    expect(triggerOrderAnswer(c, ["n", "p1", "p2"], true)).toEqual({ order: ["n", "p1", "p2"] });
  });
  it("sends the toggle on a keyed prompt", () => {
    const c = keyed();
    expect(canRememberTriggerOrder(c)).toBe(true);
    expect(triggerOrderAnswer(c, ["n", "p1", "p2"], true)).toEqual({
      order: ["n", "p1", "p2"],
      remember: true,
    });
    expect(triggerOrderAnswer(c, ["n", "p1", "p2"], false)).toEqual({
      order: ["n", "p1", "p2"],
      remember: false,
    });
  });
  it("turns the toggle off for a split order", () => {
    expect(triggerOrderAnswer(keyed(), ["p1", "n", "p2"], true)).toEqual({
      order: ["p1", "n", "p2"],
      remember: false,
    });
  });
});

describe("the notice for a remembered order", () => {
  const view = {
    seats: [{ id: "me", seat: 0, undo_auto_answer: 9 }],
    log: [],
  } as unknown as GameView;
  const entry = {
    seq: 9,
    kind: "auto_answer",
    seat: 0,
    call: "order",
    label: "Plunderer, then Noble",
    auto_answer_key: "trigger_order|noble+plunderer",
  } as unknown as LogEvent;

  it("names the order and keeps Undo and Ask me next time", () => {
    const n = autoAnswerNotice(entry, "", view, "me");
    expect(n.text).toBe("Ordered for you: Plunderer, then Noble");
    expect(n.undoable).toBe(true);
    expect(isTriggerOrderKey(n.key)).toBe(true);
    const req = autoAnswerNoticeRequest(
      n,
      () => {},
      () => {},
    );
    const ask = req.row?.find((b) => b.label === L.askMeNextTime);
    expect(ask?.title).toMatch(/Stop reusing this order/);
  });

  it("says why a remembered order was asked by hand", () => {
    expect(askedByHandText({ asked_by_hand: "changed" } as PendingChoiceView)).toMatch(/changed/);
    expect(askedByHandText({ asked_by_hand: "hidden" } as PendingChoiceView)).toMatch(/hidden/);
  });

  it("is not a trigger-order key for an ADR 0127 rule", () => {
    expect(isTriggerOrderKey("abc|triggered|draw")).toBe(false);
    expect(isTriggerOrderKey(undefined)).toBe(false);
  });
});
