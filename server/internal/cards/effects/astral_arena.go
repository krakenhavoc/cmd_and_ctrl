package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astral Arena — Plane — Kolbahan (#4870):
//
//	"No more than one creature can attack each combat.
//	 No more than one creature can block each combat.
//	 Whenever chaos ensues, Astral Arena deals 2 damage to each creature."
//
// Dueling Grounds's two lines on a plane: an AttackLimit and a whole-combat
// BlockRule.Limit, which the attack-limit gate (game.activeAttackLimitsLocked)
// and the block-rule gate both read from a face-up plane. They stop a second
// creature being declared, not one put onto the battlefield attacking or
// blocking, and count each combat separately (the rulings). The chaos damage
// comes from the plane, simultaneously, to every creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "805f9dc2-e172-481d-bae7-9a136a3d1e49",
		Name:         "Astral Arena",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackEachCombat(1)},
		BlockRules:   []game.BlockRule{NoMoreThanNCanBlockEachCombat(1)},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Astral Arena — 2 damage to each creature",
				func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item), Creature(), 2)
				}),
		},
	})
}
