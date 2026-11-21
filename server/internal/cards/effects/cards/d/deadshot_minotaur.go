package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadshot Minotaur — Creature — Minotaur {3}{R}{G}, 3/4 (Alara Reborn,
// #5586):
//
//	"When this creature enters, it deals 3 damage to target creature with
//	 flying.
//	 Cycling {R/G} ({R/G}, Discard this card: Draw a card.)"
//
// Centaur Archer's flying-only target on an enters trigger (chosen as it
// goes on the stack, CR 603.3d; removed with no flyer to target), and the
// shared Cycling row at a hybrid cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3992369a-2158-4cf8-b7e5-be517b91c1f5",
		Name:         "Deadshot Minotaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Deadshot Minotaur — it deals 3 damage to target creature with flying",
				effects.DealDamageToTheTarget(3)),
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{R/G}")},
	})
}
