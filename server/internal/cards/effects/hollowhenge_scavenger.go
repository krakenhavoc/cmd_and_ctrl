package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hollowhenge Scavenger — Creature — Elemental {3}{G}{G}, 4/5 (Innistrad, #5572):
//
//	"Morbid — When this creature enters, if a creature died this turn,
//	 you gain 5 life."
//
// Wakedancer's morbid intervening "if" (CR 603.4,
// condACreatureDiedThisTurn), checked as it enters and again as the
// trigger resolves, around a plain life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60116231-cc54-4c74-b3d9-74e517182f1f",
		Name:         "Hollowhenge Scavenger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condACreatureDiedThisTurn,
				WhenThisEnters("Hollowhenge Scavenger — morbid: you gain 5 life", Do(GainLife{Amount: 5}))),
		},
	})
}
