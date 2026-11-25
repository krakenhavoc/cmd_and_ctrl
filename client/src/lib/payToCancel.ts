// payToCancel.ts — the seat's "… unless they pay {N} before that step"
// buttons (#6220; Nafs Asp, Sabertooth Cobra).
//
// The server decides everything: `seat.pay_to_cancel` lists, for the
// viewer's own seat only, each delayed trigger waiting on them that a
// payment removes, and the button is live exactly when the viewer's own
// legal_moves carry the `special_action` move of kind `pay_to_cancel` for
// that trigger. Pressing it sends that move's params (strict and auto_tap,
// as the enumerator wrote them). NOTHING HERE IS A RULE: the title says
// why a button is greyed, and the server refuses whatever the title says.

import type { GameView, LegalMoveView, PayToCancelView } from "./protocol";

export const PAY_TO_CANCEL_KIND = "pay_to_cancel";

const STEP_WORDS: Record<string, string> = {
  upkeep: "your next upkeep",
  draw: "your next draw step",
  end: "your next end step",
};

/** The viewer's pay_to_cancel move for this payment, or null. */
export function payToCancelMove(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
  id: string,
): LegalMoveView | null {
  if (!view || !viewerID) return null;
  return (
    (view.legal_moves ?? []).find(
      (m) =>
        m.type === "special_action" &&
        m.params?.kind === PAY_TO_CANCEL_KIND &&
        m.params?.card_id === id &&
        (m.player === undefined || m.player === "" || m.player === viewerID),
    ) ?? null
  );
}

/** One payment button: what it shows and what pressing it sends. */
export interface PayToCancelButton {
  id: string;
  /** The button's text, "Pay {1}". */
  text: string;
  /** Its accessible name's argument, "{1} (Nafs Asp)". */
  what: string;
  enabled: boolean;
  /** What happens unless it is paid, and why it is greyed when it is. */
  title: string;
  /** The `special_action` params to send, when enabled. */
  params: Record<string, unknown> | null;
}

/** The viewer's payment buttons, one per waiting payment. */
export function payToCancelButtons(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
  rows: PayToCancelView[] | undefined,
): PayToCancelButton[] {
  if (!view || !viewerID || !rows) return [];
  return rows.map((row) => {
    const when = STEP_WORDS[row.at] ?? "that step";
    const what = row.source_name ? `${row.cost} (${row.source_name})` : row.cost;
    const unless = `${row.label}, at the beginning of ${when}, unless you pay ${row.cost} before then`;
    const move = payToCancelMove(view, viewerID, row.id);
    return {
      id: row.id,
      text: `Pay ${row.cost}`,
      what,
      enabled: move !== null,
      title: move ? unless : `${unless}. Pay it when you have priority and the mana`,
      params: move ? (move.params ?? null) : null,
    };
  });
}
