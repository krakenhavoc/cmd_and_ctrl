package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fae Offering — Enchantment {2}{G} (Modern Horizons 2):
//
//	"At the beginning of each end step, if you've cast both a creature spell
//	 and a noncreature spell this turn, create a Clue token, a Food token, and a
//	 Treasure token."
//
// Every player's end step, with an intervening "if" (CR 603.4) on the
// Offering's controller's own casts this turn, checked as the step begins and
// again as the trigger resolves (the per-turn cast tally).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "885d68de-db97-4c9e-bb44-8e561a482dde",
		Name:         "Fae Offering",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(faeOfferingCast,
				effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventBeginEndStep && effects.OldB033CastCreatureAndNoncreatureThisTurn(g, source.Controller)
				}, "Fae Offering — create a Clue token, a Food token, and a Treasure token",
					effects.Do(effects.CreateToken{Template: effects.ClueToken(), N: 1}, effects.CreateToken{Template: effects.FoodToken(), N: 1}, effects.CreateToken{Template: effects.TreasureToken(), N: 1}))),
		},
	})
}

var faeOfferingCast = effects.CondBoard{
	Name: "you've cast both a creature spell and a noncreature spell this turn",
	Test: func(ctx *effects.Context) bool {
		return effects.OldB033CastCreatureAndNoncreatureThisTurn(ctx.Game, ctx.Controller())
	},
}
