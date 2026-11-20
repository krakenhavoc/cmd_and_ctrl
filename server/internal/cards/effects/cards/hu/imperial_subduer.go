package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperial Subduer — Creature — Human Samurai {2}{W}, 3/2 (Kamigawa: Neon
// Dynasty):
//
//	"Whenever a Samurai or Warrior you control attacks alone, tap target
//	 creature you don't control."
//
// A targeted trigger on the lone attack (CR 506.5); the target is any
// creature its controller does not control, and it is tapped if still legal.
//
// No simplification.
func init() {
	const name = "Imperial Subduer"
	effects.Register(effects.Spec{
		OracleID:     "bedf1dd5-b133-455b-bc3d-c7b1f469d91f",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone,
					name+" — tap target creature you don't control",
					effects.B36TapChosenCreature),
				effects.TargetCreature("target creature you don't control", effects.Not(effects.YouControl()))),
		},
	})
}
