// Lobby.hint.ts — the first-use hint for creating a table (ADR 0125 §3.7).
// lobby.practice, the tutorial offer, is its own file (ADR 0125 PR 5).

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "lobby.create",
  version: 1,
  place: "lobby",
  order: 10,
  anchor: { label: L.createGame },
  when: (c) => c.signedIn,
  title: "Start a table",
  body: "Name it and create it. Its card then holds the invite link, the seats you can give to bots, and your deck.",
};

export default hint;
