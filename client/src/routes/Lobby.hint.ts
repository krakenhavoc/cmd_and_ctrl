// Lobby.hint.ts — the first-use hint for creating a table (ADR 0125 §3.7).
// lobby.practice, the tutorial offer, is its own file (ADR 0125 PR 5).
//
// Version 2 (ADR 0158 §1): the create form sits behind + New table in
// the page head, so the tip points at the button and says so.
// Version 3 (ADR 0212): a table's card is a character select. Your deck
// goes in your seat's tile and one Invite menu holds the links, so the
// old body named a button and a panel that are gone.

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "lobby.create",
  version: 3,
  place: "lobby",
  order: 10,
  anchor: { label: L.newTable },
  when: (c) => c.signedIn,
  title: "Start a table",
  body: "New table makes one. Choose your deck in your seat, then use Invite for your pod or Add bot on an open seat.",
};

export default hint;
