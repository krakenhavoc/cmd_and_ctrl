package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Let the Galaxy Burn — Sorcery {X}{5}{R} (Warhammer 40,000 Commander,
// #5948):
//
//	"Cascade
//	 Let the Galaxy Burn deals X plus 2 damage to each creature that didn't
//	 enter this turn."
//
// Cascade is the catalog keyword (it triggers on cast and resolves first,
// CR 702.85a). The damage is one instance dealt as the spell resolves to
// each creature whose current object has not entered this turn (a creature
// that entered, left and came back is a new object and is spared).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c70980b-3c38-4624-8f19-5e29d271acf5",
		Name:         "Let the Galaxy Burn",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			amount := ctx.X() + 2
			return ctx.Game.DamageInstanceForEffect(func() error {
				return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
					return !g.EnteredThisTurn(c.InstanceID)
				}), amount)
			})
		},
	})
}
