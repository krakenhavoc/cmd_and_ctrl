// PlanarZone.hint.ts — the first-use hint for a Planechase table (ADR
// 0216 §10). Offered the first time a player sits at a table that plays
// Planechase, once the first plane is face up (CR 901.5), so the card it
// points at is there to see.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";

const hint: Hint = {
  id: "table.planechase",
  version: 1,
  place: "table",
  order: 55,
  anchor: { label: L.planarZone },
  title: "Planechase",
  body: "The plane is face up. Roll the planar die on your turn for chaos, or to planeswalk.",
  when: (c) => (c.view?.planechase?.face_up.length ?? 0) > 0,
};

export default hint;
