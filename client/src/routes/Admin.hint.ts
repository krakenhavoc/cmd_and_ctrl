// Admin.hint.ts — the first-use hint for the admin views (ADR 0125 §3.7).
// Offered only in admin mode (or to the shared admin token, which is an
// admin too): the page's tabs do not render for anyone else.

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "admin.views",
  version: 1,
  place: "admin",
  order: 0,
  anchor: (c) => (c.adminMode || c.adminToken ? { label: L.adminViews } : null),
  when: (c) => c.adminMode || c.adminToken,
  title: "Admin views",
  body: "Live now, every table and every account, read from the live database. Open a row to see more.",
};

export default hint;
