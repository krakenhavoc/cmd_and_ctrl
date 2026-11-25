// payToCancel.test.ts — #6220: the seat's "… unless they pay {N} before
// that step" buttons are live exactly when the viewer's legal_moves carry
// the server's pay_to_cancel move for that payment, and send its params.

import { describe, expect, it } from "vitest";
import { payToCancelButtons } from "./payToCancel";
import type { GameView, PayToCancelView } from "./protocol";

const ME = "me";
const row: PayToCancelView = {
  id: "dt1",
  cost: "{1}",
  label: "Nafs Asp — lose 1 life",
  source_name: "Nafs Asp",
  at: "draw",
};
const move = {
  type: "special_action",
  player: ME,
  kind: "special_action",
  label: "Pay {1}: Nafs Asp — lose 1 life",
  params: { card_id: "dt1", kind: "pay_to_cancel", strict: true, auto_tap: true },
};

const viewWith = (moves: unknown[]): GameView => ({ legal_moves: moves }) as unknown as GameView;

describe("pay-to-cancel buttons (#6220)", () => {
  it("is live with the server's move, and sends its params", () => {
    const [b] = payToCancelButtons(viewWith([move]), ME, [row]);
    expect(b.enabled).toBe(true);
    expect(b.text).toBe("Pay {1}");
    expect(b.what).toBe("{1} (Nafs Asp)");
    expect(b.params).toEqual(move.params);
    expect(b.title).toContain("your next draw step");
  });

  it("is greyed without it", () => {
    const [b] = payToCancelButtons(viewWith([]), ME, [row]);
    expect(b.enabled).toBe(false);
    expect(b.params).toBeNull();
  });

  it("does not take another payment's move", () => {
    const other = { ...move, params: { ...move.params, card_id: "dt2" } };
    const [b] = payToCancelButtons(viewWith([other]), ME, [row]);
    expect(b.enabled).toBe(false);
  });

  it("shows nothing with no payment or no viewer", () => {
    expect(payToCancelButtons(viewWith([move]), ME, undefined)).toEqual([]);
    expect(payToCancelButtons(viewWith([move]), null, [row])).toEqual([]);
  });
});
