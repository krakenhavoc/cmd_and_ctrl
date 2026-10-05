// ExamplePage.hint.ts — a TEST FIXTURE for a site page's hint, with an
// anchor read off the context and a `when`. See Example.hint.ts.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";

const hint: Hint = {
  id: "admin.example",
  version: 1,
  place: "admin",
  order: 0,
  anchor: (c) => (c.adminMode || c.adminToken ? { label: L.adminViews } : null),
  when: (c) => c.adminMode || c.adminToken,
  title: "Admin views",
  body: "Live now, every table and every account. Each row links to what you can do with it.",
};

export default hint;
