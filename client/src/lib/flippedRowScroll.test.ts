// flippedRowScroll.test.ts — #4866: on an across-table (flipped)
// opponent, each battlefield row's card list is turned 180°, so what does
// not fit the list is drawn past the row's left edge. A row that scrolls
// left to right cannot reach that, and a full land strip lost its far
// lands. The rows scroll right to left instead, and what is in them still
// reads left to right. The walker and battle rows, which #4087 has never
// scroll, keep their left-to-right box.
//
// jsdom does no layout and applies no component CSS, so this pins the
// stylesheet's source. tests-e2e/tests/land-strip-reach-4866.spec.ts
// measures it in a real browser: every land on every seat's strip can be
// scrolled into view, at 2, 3 and 4 seats and three sizes.

import { describe, expect, it } from "vitest";
import panelSource from "./components/board/PlayerPanel.svelte?raw";
import rowSource from "./components/board/BattlefieldRow.svelte?raw";

const FLIPPED_ROWS = ".panel.opponent.flipped :is(.grid-creatures, .grid-middle)";

describe("a flipped opponent's rows", () => {
  it("turn their card list, inside a row that scrolls", () => {
    expect(panelSource).toContain(
      `${FLIPPED_ROWS} :global(.row-cards) {\n      transform: rotate(180deg);`,
    );
    expect(rowSource).toMatch(/\n {2}\.row \{[^}]*overflow: auto;/);
    expect(rowSource).toContain("data-zone={label}");
  });

  it("scroll right to left, so the turned list's overflow is in reach", () => {
    expect(panelSource).toContain(
      `${FLIPPED_ROWS} :global(.row[data-zone]) {\n      direction: rtl;\n    }`,
    );
  });

  it("lay out what is in them left to right, and leave the walker and battle rows alone", () => {
    // #4087: a spread row is meant never to overflow, so it keeps the
    // left-to-right scroll box it had.
    expect(panelSource).toContain(
      `${FLIPPED_ROWS} :global(.row[data-zone].spread),\n    ${FLIPPED_ROWS} :global(.row[data-zone] > *) {\n      direction: ltr;\n    }`,
    );
    // The rule this replaced put the back row's rows back to ltr, scroll
    // box and all.
    expect(panelSource).not.toMatch(/\.grid-middle > :global\(\.row\),[^{]*\{\s*direction: ltr;/);
  });
});
