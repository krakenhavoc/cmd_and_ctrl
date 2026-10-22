package effects

// Gix's Caress — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Create a tapped Powerstone token."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then the
// Powerstone (CR 111.10h) on the next line (ADR 0116 §6): the token is
// yours and cannot change what may be chosen. It enters tapped because
// this effect says so, not because Powerstones do (the 2022-10-14
// ruling). The token is made even when nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0180bed7-892a-4af9-ac4b-96c9cd4beb42",
		Name:         "Gix's Caress",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		// The token waits for the choice from the revealed hand (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(chooseNonlandFromTargetedHand),
			CreateTokenAdvanced{
				Spec: Token(PowerstoneToken()).EntersTapped(),
				N:    1,
			},
		),
	})
}
