// joinRecovery.ts — the copy for the S33 return path (#1475, per
// ADR 0044 decision 4, "seat reclaim is the backstop, not the
// mechanism", and the owner's 2026-09-24 decision 4 on #515): what
// Join.svelte tells a player who lost their session and reopened an
// old player invite on a table that has already started.
//
// Pulled out per client/README.md's testing rule ("pull the decision
// out of the component into lib/<thing>.ts and test that") — the
// notice used to say "ask your host for a spectator link" no matter
// who was asking, which was wrong for exactly this player (#520
// shipped two real ways back, and that advice was neither of them).
// The wording is the behaviour under test here, the same way
// connectionBanner.ts's copy is: a regression that quietly puts the
// old "spectator link" sentence back is the bug this file exists to
// catch.

// GUEST_RETURN_ADVICE is shown to a session-less guest: nobody can
// authenticate them into My games, so the only way back is a seat
// reclaim link only the host can mint (server/internal/lobby/reclaim.go,
// server/internal/lobby/http.go's POST …/seats/{id}/reclaim). "seat
// reclaim link" matches the wording the host's own lobby UI uses for
// the same link (Lobby.svelte's aria-label and panel header).
export const GUEST_RETURN_ADVICE =
  "Already had a seat here? Ask your host for a seat reclaim link to get back in.";

// SIGNED_IN_RETURN_LINK is where a signed-in identity (or a player
// session minted from one — myGames.ts's signedInUserID) is sent
// instead: My games, where POST /me/games/{id}/session rejoins in one
// click. #/my-games is the route router.ts parses to { name: "myGames" }.
// Shared with connectionBanner.ts's "session_ended" state (#1475, the
// owner's 2026-09-24 decision 5 on #515) — same destination, same
// reason, one constant.
export const SIGNED_IN_RETURN_LINK: { readonly href: string; readonly label: string } = {
  href: "#/my-games",
  label: "My games",
};
