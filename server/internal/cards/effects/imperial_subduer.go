package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bedf1dd5-b133-455b-bc3d-c7b1f469d91f",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventAttack, neoB09SamuraiOrWarriorAttacksAlone,
					name+" — tap target creature you don't control",
					b36TapChosenCreature),
				TargetCreature("target creature you don't control", Not(YouControl()))),
		},
	})
}
