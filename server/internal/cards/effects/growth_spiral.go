package effects

// Growth Spiral — Instant {G}{U}:
//
//	"Draw a card. You may put a land card from your hand onto the
//	 battlefield."
//
// The batch-01 first pass declared this one skipped: "a pick-from-hand
// prompt with a continuation and a hand→battlefield move" did not
// exist, and shipping the draw alone would have been a misleading
// cantrip. Both halves exist now — `PutFromHandOntoBattlefield` is the
// prompt and `game.PutFromHandOntoBattlefieldForEffect` is the move —
// and `MayPutALandFromHand` is the printed clause five catalog cards
// already share, this one included by name in its doc comment.
//
// The land is PUT, not played: it does not use the land drop
// (CR 305.2 is about playing one), which is the entire reason the
// card is an instant-speed ramp spell rather than a cantrip. It enters
// untapped, and its own enters-tapped replacement still runs, so a
// shockland or a slowland behaves exactly as it would off a land drop.
//
// Ordering: the draw happens first and the prompt is raised over the
// hand the draw left, so a land drawn off the top is a legal pick —
// which is printed behaviour and is why the two sentences are in that
// order. The prompt is optional, so declining is a real answer and a
// player with no land in hand is never asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34bcc217-dd91-45a0-90d7-a94d02f1f317",
		Name:         "Growth Spiral",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The land offer waits for a prompt the draw asks (CR 608.2c, #4155).
			DrawCards{N: 1},
			MayPutALandFromHand("Growth Spiral"),
		),
	})
}
