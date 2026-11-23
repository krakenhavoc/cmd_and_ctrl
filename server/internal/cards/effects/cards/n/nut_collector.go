package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nut Collector — Creature — Human Druid {5}{G}, 1/1 (Dominaria
// Remastered):
//
//	"At the beginning of your upkeep, you may create a 1/1 green Squirrel
//	 creature token.
//	 Threshold — All Squirrels get +2/+2 as long as there are seven or more
//	 cards in your graveyard."
//
// The upkeep token is a "you may" trigger over Squirrel creation. The
// threshold anthem is a layer 7c static over EVERY Squirrel, whoever
// controls it (a TribalAnthem with no controller clause), gated on the
// Collector's controller's graveyard holding seven or more cards
// (FraThresholdMet), read on every recompute. The Collector is not itself a
// Squirrel and gets nothing.
//
// No simplification.
func init() {
	filter := effects.TribeFilter{Tribes: []string{"Squirrel"}}
	anthem := effects.TribalAnthem(filter, 2, 2)
	anthem.Label = "Threshold — All Squirrels get +2/+2 as long as there are seven or more cards in your graveyard."
	applies := anthem.AppliesTo
	anthem.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.FraThresholdMet(g, source.Controller) && applies(target, g, source)
	}
	effects.Register(effects.Spec{
		OracleID:     "4e1bc03c-e131-4422-beba-0a45a26fdf43",
		Name:         "Nut Collector",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Nut Collector — create a 1/1 green Squirrel creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Squirrel"), N: 1})),
				"Nut Collector — create a 1/1 green Squirrel token?"),
		},
		Static: []game.StaticAbility{anthem},
	})
}
