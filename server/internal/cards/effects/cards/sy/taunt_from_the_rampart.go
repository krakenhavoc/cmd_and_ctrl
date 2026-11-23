package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Taunt from the Rampart — Sorcery {3}{R}{W} (Marvel Super Heroes
// Commander, #4455):
//
//	"Goad all creatures your opponents control. Until your next turn, those
//	 creatures can't block."
//
// Disrupt Decorum's goad over every creature you don't control, then the
// same set (locked as the spell resolves, CR 611.2c) gets a can't-block
// restriction until your next turn, the registration Detain makes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9138a768-d01d-4603-9bae-d3bff6cd2f61",
		Name:         "Taunt from the Rampart",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			theirs := effects.And(effects.Creature(), effects.OpponentControls())
			if err := effects.GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
				return c.Controller != ctx.Controller()
			}); err != nil {
				return err
			}
			set := effects.EotSnapshot(ctx, uuid.Nil, theirs)
			if set == nil {
				return nil
			}
			ctx.Game.RegisterScopedEffectForEffect(ctx.Source(), set.AffectedObjects(),
				[]game.Mod{game.AddRestrictionsMod(game.CantBlock)},
				effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
				"Taunt from the Rampart — those creatures can't block")
			return nil
		},
	})
}
