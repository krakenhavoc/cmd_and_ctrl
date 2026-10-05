// LobbyPractice.hint.ts — the tutorial offer (ADR 0125 §3.7, §6; ADR
// 0076 §2.6's first-visit offer). Offered once on the Lobby to a guest,
// or to a signed-in person who has not finished a game, and never to
// the admin token. Its action opens a practice table, and "Got it"
// reads "Not now" beside it.
//
// Whether a signed-in person has finished a game comes from the Lobby's
// own read of GET /me/games (lib/hints/endedGame.ts). Until that read
// lands the anchor names nothing, so the visit waits for it (up to the
// site window) instead of showing the offer to someone it is not for,
// or passing over it for a later hint.

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "lobby.practice",
  version: 1,
  place: "lobby",
  order: 0,
  anchor: (c) => (c.signedIn && c.hasEndedGame === null ? null : { label: L.lobbyTitle }),
  when: (c) => !c.adminToken && (!c.signedIn || c.hasEndedGame !== true),
  title: "New here?",
  body: "A five-minute practice game against a bot shows you where everything is.",
  action: { label: L.startPractice, href: "#/practice" },
};

export default hint;
