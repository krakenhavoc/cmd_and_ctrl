// AttackDeclarationModal.hint.ts — the first-use hint for exert (ADR
// 0130 §7, ADR 0125 §7). Offered once the viewer controls a creature
// that may be exerted as it attacks: every attack path asks about it,
// the dock's click, the card menu and this picker's Exert toggle.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { controlsExertCreature } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.exert",
  version: 1,
  place: "table",
  order: 45,
  anchor: { label: L.actions },
  title: "Exerting an attacker",
  body: "When you attack with it, choose Attack or Attack and exert. An exerted creature won't untap during your next untap step.",
  when: controlsExertCreature,
};

export default hint;
