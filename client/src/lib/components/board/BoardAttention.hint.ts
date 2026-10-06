// BoardAttention.hint.ts — the first-use hint for the attention strip
// (ADR 0125 §3.7). The strip is always mounted and has no area while it
// is empty, so the engine offers this hint only once something is in it.
// An item on the stack is also drawn there in the compact style, and
// that is not news, so the hint waits for an empty stack.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { stackIsEmpty } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.attention",
  version: 1,
  place: "table",
  order: 60,
  anchor: { label: L.attention },
  title: "News, top left",
  body: "What the bots are doing, cards revealed to you and the roll call show up here. None of it waits on you.",
  when: stackIsEmpty,
  // Empty, the strip has no area: wait for something in it, unlogged.
  waitsForAnchor: true,
};

export default hint;
