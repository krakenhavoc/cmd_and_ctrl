// Settings.hint.ts — the first-use hint for the Settings dialog (ADR 0125
// §3.7). Place "settings": it draws inside the dialog, through HintSlot.

import { L } from "../labels";
import type { Hint } from "../hints/hint";

const hint: Hint = {
  id: "settings.display",
  version: 1,
  place: "settings",
  order: 0,
  anchor: { label: L.settingsSections },
  title: "Skins are under Display",
  body: "Pick a skin and an accent colour there. Signed in, most settings follow you to your other browsers.",
};

export default hint;
