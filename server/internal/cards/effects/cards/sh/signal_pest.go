package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Signal Pest — Artifact Creature — Pest {1}, 0/1 (Mirrodin Besieged,
// #5576):
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)
//	 This creature can't be blocked except by creatures with flying or
//	 reach."
//
// Battle cry (BattleCry, CR 702.91a) and Silhana Ledgewalker's block rule
// widened to Canopy Cover's "creatures with flying or reach"
// (CantBeBlockedExceptBy), checked as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "724f3d2f-0513-4efd-961f-56a7ea7b5731",
		Name:         "Signal Pest",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BattleCry("Signal Pest")},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.Or(effects.HasKeyword("flying"), effects.HasKeyword("reach")), "creatures with flying or reach"),
		},
	})
}
