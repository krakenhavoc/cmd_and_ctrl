package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forge Devil — Creature — Devil {R}, 1/1 (Dark Ascension, #5571):
//
//	"When this creature enters, it deals 1 damage to target creature
//	 and 1 damage to you."
//
// Spicy Oatmeal Pizza's body (tmtB6DamageAnyTargetAndYou) with 1 and 1 and
// a creature target: both are dealt at once, as one damage instance. With
// its only target gone the trigger does nothing, and you take no damage
// (CR 608.2b). The Devil itself is a legal target, and the only one on an
// otherwise empty board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be7a38a3-c27f-4917-a59c-763d0b36e427",
		Name:         "Forge Devil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Forge Devil — 1 damage to target creature and 1 damage to you",
				tmtB6DamageAnyTargetAndYou(1, 1)), TargetCreature("target creature")),
		},
	})
}
