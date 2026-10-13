// Lobby.hint.ts — the first-use hint for creating a table (ADR 0125 §3.7).
// lobby.practice, the tutorial offer, is its own file (ADR 0125 PR 5).
//
// Version 2 (ADR 0158 §1): the create form sits behind + New table in
// the page head, and a table's card holds Copy invite, its seats and a
// ⋯ menu, so the tip points at the button and says so.

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "lobby.create",
  version: 2,
  place: "lobby",
  order: 10,
  anchor: { label: L.newTable },
  when: (c) => c.signedIn,
  title: "Start a table",
  body: "New table makes one. Its card then has Copy invite for your pod, Add bot on the empty seats, and your deck underneath.",
};

export default hint;
