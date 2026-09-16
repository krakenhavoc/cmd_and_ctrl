import { describe, expect, it } from "vitest";
import { rejectionForPrompt, type ChoiceSubmission, type ServerErrorLike } from "./choiceRejection";

// #624: a choose_cards answer that breaks the card's set-level rule is
// refused with the prompt left open. The board's toast is under the
// modal's backdrop, so ChoicePromptModal shows the refusal itself, and
// this decides when an error is the modal's to show.

const RULE_MESSAGE =
  "that selection doesn't meet the card's condition — check its text and choose again";

function sent(choiceID: string, sentAt: number): ChoiceSubmission {
  return { choiceID, sentAt };
}

function error(atMs: number, message = RULE_MESSAGE, code = "bad_request"): ServerErrorLike {
  return { code, message, at: new Date(atMs) };
}

describe("rejectionForPrompt", () => {
  it("shows the server's refusal of the open prompt's own answer", () => {
    expect(rejectionForPrompt(sent("c1", 1_000), "c1", error(1_050))).toEqual({
      code: "bad_request",
      message: RULE_MESSAGE,
    });
  });

  it("counts a reply stamped in the same millisecond as the answer", () => {
    expect(rejectionForPrompt(sent("c1", 1_000), "c1", error(1_000))).not.toBeNull();
  });

  it("shows nothing when there is no error", () => {
    expect(rejectionForPrompt(sent("c1", 1_000), "c1", null)).toBeNull();
  });

  // The prompt opened with an older toast still up (a rejected cast a
  // moment before). The player has not answered yet, so that error is
  // not about this prompt.
  it("does not claim an error when the modal has sent nothing", () => {
    expect(rejectionForPrompt(null, "c1", error(1_050))).toBeNull();
  });

  it("does not claim an error that predates the answer", () => {
    expect(rejectionForPrompt(sent("c1", 1_000), "c1", error(999))).toBeNull();
  });

  // The answer was accepted and the next prompt in the queue opened;
  // an unrelated error arriving now is not the new prompt's refusal.
  it("does not carry a refusal onto a different prompt", () => {
    expect(rejectionForPrompt(sent("c1", 1_000), "c2", error(1_050))).toBeNull();
  });

  it("shows nothing once the prompt has closed", () => {
    expect(rejectionForPrompt(sent("c1", 1_000), null, error(1_050))).toBeNull();
  });
});
