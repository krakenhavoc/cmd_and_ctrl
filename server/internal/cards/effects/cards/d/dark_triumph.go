package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Triumph — Instant {4}{B}
// (Nemesis, #5618):
//
//	"If you control a Swamp, you may sacrifice a creature rather than pay this spell's mana cost.
//	 Creatures you control get +2/+0 until end of turn."
//
// nemSacrificeACreatureWithASwamp on Red Alert!'s body: "creatures you
// control" is read once, as the spell resolves (CR 611.2c), so a
// creature sacrificed for the cost gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "885630a8-c43b-4804-acd6-14a7c7321d73",
		Name:             "Dark Triumph",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.NemSacrificeACreatureWithASwamp()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "Dark Triumph — creatures you control get +2/+0"}.Apply(ctx)
		},
	})
}
