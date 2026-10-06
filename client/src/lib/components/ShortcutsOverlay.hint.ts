// ShortcutsOverlay.hint.ts — the first-use hint for the keyboard
// shortcuts (ADR 0125 §3.7). It names the player's own keys, offered on
// a device with a hover-capable fine pointer from the viewer's second
// turn. It is ordered after `table.dock` and `table.more`, so the dock's
// hint has always been shown first.

import { L } from "../labels";
import type { Hint } from "../hints/hint";
import { atViewersSecondTurn, hasHoverPointer } from "../hints/tableWhen";

const hint: Hint = {
  id: "table.shortcuts",
  version: 1,
  place: "table",
  order: 90,
  anchor: { label: L.actions },
  title: "Keys for everything",
  body: (k) => {
    if (!k.helpKey) return "Every keyboard shortcut is listed under Shortcuts in Settings.";
    const next = k.nextKey ? ` ${k.nextKey} presses next.` : "";
    return `Press ${k.helpKey} to see every keyboard shortcut.${next}`;
  },
  when: (c) => hasHoverPointer() && atViewersSecondTurn(c),
};

export default hint;
