// blockTax.ts — the CR 509.1d block tax on block staging (#5375,
// ADR 0221). Archangel of Tithes attacking: "creatures can't block
// unless their controller pays {1} for each of those creatures".
//
// The price is the server's (`turn.block_taxes`); nothing here prices a
// block. It is paid as the declaration completes, so the staging dock
// shows what finishing will cost and finish_blocks sends `auto_tap`.

import type { BlockTaxView, GameView } from "./protocol";

// blockTaxFor is the viewer's row of `turn.block_taxes`, or null when
// their blocks cost nothing.
export function blockTaxFor(
  view: GameView | null | undefined,
  seatID: string | null,
): BlockTaxView | null {
  if (!view || !seatID) return null;
  const rows = view.turn?.block_taxes;
  if (!rows || rows.length === 0) return null;
  const idx = view.seats?.findIndex((s) => s.id === seatID) ?? -1;
  if (idx < 0) return null;
  return rows.find((r) => r.seat === idx) ?? null;
}

// finishBlocksParams is finish_blocks' payload: auto_tap when the
// staged blocks are taxed, so the server taps lands for the price (the
// attack tax's one-click posture), and nothing otherwise.
export function finishBlocksParams(tax: BlockTaxView | null): { auto_tap: true } | undefined {
  return tax?.cost ? { auto_tap: true } : undefined;
}

// blockTaxNote is the dock's sentence: what finishing costs now, or,
// with nothing taxed staged, what each blocker will cost.
export function blockTaxNote(tax: BlockTaxView | null): string {
  if (!tax) return "";
  if (tax.cost) return `blocking costs ${tax.cost}, paid with your lands when you finish`;
  if (tax.each) return `each blocker costs ${tax.each}`;
  return "blocking has a cost";
}
