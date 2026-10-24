import { describe, expect, it } from "vitest";
import type { StickerSheetView } from "./protocol";
import {
  freeStickerCount,
  sheetsTitle,
  stickersTitle,
  stickerText,
  ticketCostText,
} from "./stickers";

// ADR 0208 (#2138): the words for stickers and sheets.

const sheet: StickerSheetView = {
  name: "Night Brushwagg Ringmaster",
  stickers: [
    { kind: "name", text: "Night" },
    { kind: "ability", text: "Menace", ticket_cost: 2, in_use: true },
    { kind: "pt", text: "2/3", ticket_cost: 2 },
    { kind: "art" },
  ],
};

describe("stickers", () => {
  it("phrases each kind", () => {
    expect(stickerText({ kind: "name", text: "Night" })).toBe("“Night”");
    expect(stickerText({ kind: "ability", text: "Menace" })).toBe("Menace");
    expect(stickerText({ kind: "pt", text: "2/3" })).toBe("2/3");
    expect(stickerText({ kind: "art" })).toBe("art sticker");
  });

  it("says a ticket cost", () => {
    expect(ticketCostText(0)).toBe("");
    expect(ticketCostText(undefined)).toBe("");
    expect(ticketCostText(1)).toBe("1 ticket");
    expect(ticketCostText(6)).toBe("6 tickets");
  });

  it("lists an object's stickers", () => {
    expect(stickersTitle([{ kind: "ability", text: "Persist" }, { kind: "art" }])).toBe(
      "sticker: Persist\nsticker: art sticker",
    );
  });

  it("counts and lists only the free stickers of a seat's sheets", () => {
    expect(freeStickerCount([sheet])).toBe(3);
    expect(sheetsTitle([sheet])).toBe(
      "Night Brushwagg Ringmaster\n  “Night”\n  2/3 (2 tickets)\n  art sticker",
    );
  });
});
