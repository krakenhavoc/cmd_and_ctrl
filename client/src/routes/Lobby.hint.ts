// Lobby.hint.ts — the first-use hint for creating a table (ADR 0125 §3.7).
// lobby.practice, the tutorial offer, is its own file (ADR 0125 PR 5).
//
// Version 2 (ADR 0158 §1): the create form sits behind + New table in
// the page head, so the tip points at the button and says so. A table's
// card holds its seats and, since #3499, a toolbar with Copy invite and
// your deck; the body still reads true, so the version stands.

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
