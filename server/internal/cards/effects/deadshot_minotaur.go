package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3992369a-2158-4cf8-b7e5-be517b91c1f5",
		Name:         "Deadshot Minotaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Deadshot Minotaur — it deals 3 damage to target creature with flying",
				DealDamageToTheTarget(3)),
				TargetCreature("target creature with flying", HasKeyword("flying"))),
		},
		Activated: []ActivatedAbility{Cycling("{R/G}")},
	})
}
