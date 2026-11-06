package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arabella, Abandoned Doll — Legendary Artifact Creature — Toy {R}{W},
// 1/3 (Duskmourn: House of Horror, #4466):
//
//	"Whenever Arabella attacks, it deals X damage to each opponent and
//	 you gain X life, where X is the number of creatures you control
//	 with power 2 or less."
//
// X is counted once, as the trigger resolves, over the creatures its
// controller controls then (Arabella included while her power is 2 or
// less). The damage is Colossus of the Blood Age's drain shape: one
// event of X damage from Arabella to every opponent, then X life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c4bbb1b-29c4-4e06-aed0-b361293a585b",
		Name:         "Arabella, Abandoned Doll",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Arabella, Abandoned Doll — X damage to each opponent and you gain X life", arabellaDrain),
		},
	})
}

// arabellaDrain counts the creatures with power 2 or less and drains by
// that many.
func arabellaDrain(g *game.Game, item *game.StackItem) error {
	x := countControlled(g, item.Controller, func(c game.Card) bool {
		return c.IsCreature() && c.CurrentPower() <= 2
	})
	if x <= 0 {
		return nil
	}
	return Do(DamageEachOpponent{N: x}, GainLife{Amount: x})(g, item)
}
