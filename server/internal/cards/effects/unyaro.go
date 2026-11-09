package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unyaro — Plane — Zhalfir (#4870):
//
//	"At the beginning of your end step, if you planeswalked to Unyaro this
//	 turn, untap all creatures. They phase out until a player planeswalks.
//	 (Treat them and anything attached to them as though they didn't exist.)
//	 Whenever chaos ensues, create two 2/2 white and blue Knight creature
//	 tokens with vigilance."
//
// The end-step ability has an intervening "if" (CR 603.4), checked as it
// triggers and again as it resolves against the turn's planeswalk events. The
// creatures phase out until the plane stops being face up
// (game.PhaseOutUntilPlaneswalkForEffect): they skip their controller's untap
// step phase-in, as the first ruling says, and phase in the moment a player
// planeswalks. Phasing out is not leaving the battlefield (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "503fcc43-dafd-484b-a97f-bbd11eed66de",
		Name:         "Unyaro",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(b6PlaneswalkedToThisPlaneThisTurn{},
				AtYourEndStep("Unyaro — untap all creatures; they phase out until a player planeswalks",
					b6UntapThenPhaseOutAllCreaturesUntilPlaneswalk)),
			WheneverChaosEnsues("Unyaro — create two 2/2 white and blue Knight creature tokens with vigilance",
				Do(CreateToken{Template: TokenCard("2/2 white and blue Knight with vigilance"), N: 2})),
		},
	})
}
