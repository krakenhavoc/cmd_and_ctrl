package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "60116231-cc54-4c74-b3d9-74e517182f1f",
		Name:         "Hollowhenge Scavenger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn,
				effects.WhenThisEnters("Hollowhenge Scavenger — morbid: you gain 5 life", effects.Do(effects.GainLife{Amount: 5}))),
		},
	})
}
