package effects

// Eliminate the Impossible — Instant {1}{U}:
//
//	"Investigate. Creatures your opponents control get -2/-0 until end
//	 of turn. If any of them are suspected, they're no longer suspected.
//	 (To investigate, create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// The -2/-0 reaches the creatures your opponents control as the spell
// resolves and no others (CR 611.2c), and "any of them" is the same
// set: UnsuspectAll with the same match, so a suspected creature you
// control is left alone and one an opponent gains control of later this
// turn is not retroactively included.
//
// No simplification.
func init() {
	opponentsCreatures := func() CardPredicate { return And(Creature(), OpponentControls()) }
	Register(Spec{
		OracleID:     "c78d4cec-6764-4cce-96d2-a2d85a57b218",
		Name:         "Eliminate the Impossible",
		Completeness: CompletenessFull,
		// Each sentence waits for the prompt the one before asks (a CR 616 order on the Clue;
		// CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			CreateToken{Template: ClueToken(), N: 1},
			BoostUntilEOT{
				Match: opponentsCreatures(), Power: -2,
				Label: "Eliminate the Impossible — -2/-0 until end of turn",
			},
			UnsuspectAll{Match: opponentsCreatures()},
		),
	})
}
