// choiceRejection decides which server error, if any, belongs inside
// the open ChoicePromptModal (#624).
//
// A resolve_choice the server refuses leaves the prompt in
// pending_choices, so the modal stays open and the player is meant to
// choose again. The only other place an error frame shows is the
// attention-strip toast on the board, and the modal's full-screen
// backdrop (z-index 200, dimmed and blurred) sits on top of it. So a
// rejection the player can fix, such as "discard two cards unless you
// discard a creature card" answered with one land, has to be shown in
// the prompt itself or it is not shown at all.
//
// The client's lastError store is global: it holds whatever the
// server last refused, from any action. The modal claims an error only
// when it plainly answers the modal's own submit: the player sent an
// answer to THIS prompt, the prompt is still the one open, and the
// error arrived after that answer went out. Anything older, or an
// error while a different prompt is showing, stays the toast's.

/** The modal's record of the last answer it sent. */
export interface ChoiceSubmission {
  choiceID: string;
  /** Date.now() when the answer was sent. */
  sentAt: number;
}

/** The fields of GameClient.lastError this decision reads. */
export interface ServerErrorLike {
  code: string;
  message: string;
  at: Date;
}

/** What the modal shows. */
export interface ChoiceRejection {
  code: string;
  message: string;
}

/**
 * rejectionForPrompt returns the error to show inside the open prompt,
 * or null when the current error (if any) is not a reply to the
 * prompt's own last answer.
 */
export function rejectionForPrompt(
  submission: ChoiceSubmission | null,
  activeChoiceID: string | null,
  err: ServerErrorLike | null,
): ChoiceRejection | null {
  if (!submission || !err || activeChoiceID === null) return null;
  if (submission.choiceID !== activeChoiceID) return null;
  // Same clock on both sides: sentAt is stamped before the frame goes
  // out and `at` when the reply is received, so a reply is never
  // earlier than its request. An equal millisecond is still a reply.
  if (err.at.getTime() < submission.sentAt) return null;
  return { code: err.code, message: err.message };
}
