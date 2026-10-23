import { describe, expect, it } from "vitest";
import { isInlineChoice, payPhrase } from "./choiceDock";
import {
  branchLabel,
  branchNamesCards,
  branchPicker,
  branchReady,
  branchWhyNot,
  defaultPickerBranch,
} from "./payEither";
import type { PayOptionView, PendingChoiceView } from "./protocol";
import { payUnlessAnswer } from "./waterbend";

// ADR 0187 §5 (#3575): an either payment, Worms of the Earth.

const sacrifice: PayOptionView = {
  action: "sacrifice",
  count: 2,
  words: "sacrifice two lands",
  payable: true,
  options: ["l1", "l2", "l3"],
};
const damage: PayOptionView = { action: "damage", count: 5, words: "take 5 damage", payable: true };

const choice = {
  id: "c1",
  kind: "pay_unless",
  pay_cost: "sacrifice two lands or take 5 damage",
  pay_either: [sacrifice, damage],
} as PendingChoiceView;

describe("the answer to an either payment names the branch", () => {
  it("damage is { apply: true, pay_option: 1 } with no cards", () => {
    expect(payUnlessAnswer(choice, true, [], [], 0, 1)).toEqual({
      choice_id: "c1",
      apply: true,
      pay_option: 1,
    });
  });

  it("the sacrifice is { apply: true, pay_option: 0, card_ids }", () => {
    expect(payUnlessAnswer(choice, true, [], ["l1", "l3"], 0, 0)).toEqual({
      choice_id: "c1",
      apply: true,
      pay_option: 0,
      card_ids: ["l1", "l3"],
    });
  });

  it("the first branch is index 0, which is sent", () => {
    expect(payUnlessAnswer(choice, true, [], [], 0, 0)).toHaveProperty("pay_option", 0);
  });

  it("a decline never names a branch or cards", () => {
    expect(payUnlessAnswer(choice, false, [], ["l1"], 0, 1)).toEqual({
      choice_id: "c1",
      apply: false,
    });
  });

  it("a plain payment never sends pay_option", () => {
    const plain = { id: "c2", kind: "pay_unless", pay_cost: "{2}" } as PendingChoiceView;
    expect(payUnlessAnswer(plain, true, [], [], 0, 1)).toEqual({ choice_id: "c2", apply: true });
  });
});

describe("branches", () => {
  it("name cards or don't", () => {
    expect(branchNamesCards(sacrifice)).toBe(true);
    expect(branchNamesCards(damage)).toBe(false);
    expect(branchNamesCards({ ...sacrifice, action: "exile_graveyard" })).toBe(true);
    expect(branchNamesCards({ ...sacrifice, action: "discard" })).toBe(true);
    expect(branchNamesCards({ ...sacrifice, action: "tap" })).toBe(true);
  });

  it("read as buttons", () => {
    expect(branchLabel(sacrifice)).toBe("Sacrifice two lands");
    expect(branchLabel(damage)).toBe("Take 5 damage");
  });

  it("a card branch is ready on exactly its count, a names-nothing branch at once", () => {
    expect(branchReady(sacrifice, ["l1"])).toBe(false);
    expect(branchReady(sacrifice, ["l1", "l2"])).toBe(true);
    expect(branchReady(sacrifice, ["l1", "l2", "l3"])).toBe(false);
    expect(branchReady(damage, [])).toBe(true);
  });

  it("an unpayable branch is never ready, and says why", () => {
    const short = { ...sacrifice, payable: false, options: undefined };
    expect(branchReady(short, [])).toBe(false);
    expect(branchWhyNot(short)).toContain("sacrifice two lands");
    expect(branchWhyNot(damage)).toBeUndefined();
  });

  it("feed the shared card picker", () => {
    expect(branchPicker(sacrifice)).toEqual({
      action: "sacrifice",
      count: 2,
      options: ["l1", "l2", "l3"],
    });
    expect(branchPicker({ ...sacrifice, options: undefined }).options).toEqual([]);
  });

  it("open the first payable card branch", () => {
    expect(defaultPickerBranch([damage, sacrifice])).toBe(1);
    expect(defaultPickerBranch([{ ...sacrifice, payable: false }, damage])).toBe(-1);
    expect(defaultPickerBranch([damage, damage])).toBe(-1);
  });
});

describe("the prompt's words", () => {
  it("read the whole offer as one verb phrase", () => {
    expect(payPhrase(choice)).toBe("sacrifice two lands or take 5 damage");
  });

  it("is never answered inline: it has a button for each way to pay", () => {
    expect(isInlineChoice(choice)).toBe(false);
  });
});
