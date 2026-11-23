package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tangled Colony — Creature — Rat {1}{B}, 2/2 (Wilds of Eldraine, #4471):
//
//	"This creature can't block.
//	 When this creature dies, create X 1/1 black Rat creature tokens with
//	 "This token can't block," where X is the amount of damage dealt to
//	 it this turn."
//
// "Can't block" is Carrion Feeder's RestrictSelf. X is the damage marked
// on the Colony as it last existed (game.PermanentInfo.DamageMarked, the
// departure record, CR 608.2h), so damage that killed it and damage from
// earlier in the turn both count, and a Colony that died with none makes
// nothing. The Rat is Harried Spearguard's catalog token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f01c327-ba50-4fc3-a930-fa1fb02e34b9",
		Name:         "Tangled Colony",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tangled Colony — create X 1/1 black Rat creature tokens with \"This token can't block,\" where X is the amount of damage dealt to it this turn",
				func(g *game.Game, item *game.StackItem) error {
					info, ok := g.LastKnownPermanentForEffect(item.SourceCardID)
					if !ok || info.DamageMarked <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.CantBlockRatToken(), N: info.DamageMarked}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
