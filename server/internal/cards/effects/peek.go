package effects

// Peek — Instant {U} (EDHREC rank 4480):
//
//	"Look at target player's hand.
//	 Draw a card."
//
// Gitaxian Probe with a mana cost and instant speed. One blue for a
// cantrip that replaces itself is already free; the information is
// the reason a control deck runs it — knowing whether the counter is
// live before you tap out is worth a card that costs you nothing.
//
// The body is lookAtTargetPlayersHandThenDraw, shared with Gitaxian
// Probe: "look at" is not "reveal", so only the CASTER learns the
// hand. See that helper for why, and for the resolution-time read.
//
// A target player who left the game in response fizzles the spell and
// you draw nothing (CR 608.2b); a player with an empty hand is still
// a legal target and you still draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "873abbf4-c9b5-4b8f-8cd0-613cf3b9b1d5",
		Name:         "Peek",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    lookAtTargetPlayersHandThenDraw,
	})
}
