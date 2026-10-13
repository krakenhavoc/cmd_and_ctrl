// HelpMenu.hint.ts — the tip about tips (ADR 0125 §3.7 `site.help`,
// §6). A "site" hint, so any site page offers it once that page's own
// hints are seen or do not apply. The front page is the sign-in page
// and offers no tips (ADR 0125 §3.8, ADR 0156 §4).

import { L } from "../labels";
import type { Hint } from "../hints/hint";

const hint: Hint = {
  id: "site.help",
  version: 1,
  place: "site",
  order: 0,
  anchor: { label: L.help },
  title: "Tips show once",
  body: "Each page shows a tip like this the first time you visit. Help, up here, shows them again and starts a practice game.",
};

export default hint;
