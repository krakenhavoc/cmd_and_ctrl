// stickers.ts — the words the table shows for stickers and sticker sheets
// (ADR 0208, #2138). The server says what each sticker is; this only
// phrases it.

import type { StickerSheetView, StickerView } from "./protocol";

/** One sticker in a few words: “Night”, Menace, 2/3, art sticker. */
export function stickerText(s: StickerView): string {
  switch (s.kind) {
    case "name":
      return `“${s.text ?? ""}”`;
    case "art":
      return "art sticker";
    default:
      return s.text ?? "";
  }
}

/** A ticket cost, “2 tickets”, or "" when it has none. */
export function ticketCostText(n: number | undefined): string {
  if (!n) return "";
  return n === 1 ? "1 ticket" : `${n} tickets`;
}

/** The hover text for the stickers on an object, one per line. */
export function stickersTitle(stickers: readonly StickerView[]): string {
  return stickers.map((s) => `sticker: ${stickerText(s)}`).join("\n");
}

/** How many of a seat's stickers are free to put on something. */
export function freeStickerCount(sheets: readonly StickerSheetView[]): number {
  let n = 0;
  for (const sheet of sheets) {
    for (const s of sheet.stickers) {
      if (!s.in_use) n++;
    }
  }
  return n;
}

/**
 * The hover text for a seat's sheets: each sheet, then its free stickers
 * with their ticket costs. A sticker that is on an object is left out.
 */
export function sheetsTitle(sheets: readonly StickerSheetView[]): string {
  const lines: string[] = [];
  for (const sheet of sheets) {
    lines.push(sheet.name);
    for (const s of sheet.stickers) {
      if (s.in_use) continue;
      const cost = ticketCostText(s.ticket_cost);
      lines.push(`  ${stickerText(s)}${cost ? ` (${cost})` : ""}`);
    }
  }
  return lines.join("\n");
}
