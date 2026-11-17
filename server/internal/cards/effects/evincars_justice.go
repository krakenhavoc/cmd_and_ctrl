package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evincar's Justice — Sorcery {2}{B}{B} (Tempest, #5628):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Evincar's Justice deals 2 damage to each creature and each player."
//
// Inferno's body at 2 (b23DamageEachCreatureAndEachPlayer: every
// creature, then every player still in the game, its caster included,
// as one damage event), with Capsize's buyback (CR 702.27a). With no
// targets it always resolves, so a paid buyback always returns it to its
// owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c0d3dfc0-1ff2-4ac3-98eb-32eaa5b84c05",
		Name:          "Evincar's Justice",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, 2)
		},
	})
}
