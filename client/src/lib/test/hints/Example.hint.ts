// Example.hint.ts — a TEST FIXTURE, not a hint the client ships (ADR 0125
// Delivery PR 4). lib/hints/index.ts excludes lib/test/ from its glob;
// hints.test.ts, queue.test.ts and labels.test.ts collect this file on
// their own to prove the collection, the rules and the anchor walk.
// The real hints arrive with ADR 0125's PRs 5–7, beside their features.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";

const hint: Hint = {
  id: "table.example",
  version: 1,
  place: "table",
  order: 0,
  anchor: { label: L.actions },
  title: "Your controls",
  body: (k) =>
    k.nextKey
      ? `The next button (${k.nextKey}) moves the game on, and anything you need to answer opens here.`
      : "The next button moves the game on, and anything you need to answer opens here.",
};

export default hint;
