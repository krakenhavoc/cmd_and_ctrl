package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molecular Modifier — Creature — Kavu Artificer {2}{R}, 2/2 (Edge of
// Eternities, #4461):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gets +1/+0 and gains first strike until end of turn."
//
// Hog-Monkey's beginning-of-combat trigger with Belligerent of the
// Ball's body (firstTargetGainsUntilEndOfTurn): +1/+0 and first strike on
// the target until end of turn. The Modifier can target itself.
//
// No simplification.
func init() {
	const label = "Molecular Modifier — target creature you control gets +1/+0 and gains first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "d0e58b8e-2c41-4602-a905-c2a6086ecf70",
		Name:         "Molecular Modifier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(label, effects.FirstTargetGainsUntilEndOfTurn(1, label, "first strike")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
