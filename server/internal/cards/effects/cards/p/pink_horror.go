package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pink Horror — Creature — Demon Horror {3}{U}{R}, 4/4 (Warhammer 40,000
// Commander, S82 GAP batch 194, #5948):
//
//	"Coruscating Flames — Whenever you cast an instant or sorcery spell, this
//	 creature deals 2 damage to any target.
//	 Split — When this creature dies, create two 2/2 blue and red Demon Horror
//	 creature tokens named Blue Horror with "Whenever you cast an instant or
//	 sorcery spell, this token deals 1 damage to any target.""
//
// Two triggers. The cast trigger chooses its target as it goes on the stack.
// The dies trigger makes the tokens for the controller as the Pink Horror last
// was (CR 603.10a); the Blue Horror is the printed token, whose own cast
// trigger lives on it so its damage is the token's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3aa4c374-b146-4890-a8e1-717f214d9741",
		Name:         "Pink Horror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Pink Horror — 2 damage to any target", effects.DamageToFirstTarget(2)),
				effects.TargetAny()),
			effects.WhenThisDies("Pink Horror — create two 2/2 blue and red Demon Horror creature tokens named Blue Horror",
				effects.Do(effects.CreateToken{Template: effects.GapB194BlueHorrorToken(), N: 2})),
		},
	})
}
