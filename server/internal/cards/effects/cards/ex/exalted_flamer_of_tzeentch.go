package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exalted Flamer of Tzeentch — "At the beginning of your upkeep, return an
// instant or sorcery card at random from your graveyard to your hand.
// Whenever you cast an instant or sorcery spell, this creature deals
// 1 damage to each opponent."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6591a589-deb4-462f-a4f1-f6261f3433c3",
		Name:         "Exalted Flamer of Tzeentch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery())), "Exalted Flamer of Tzeentch — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			}),
			effects.AtYourUpkeep("Exalted Flamer of Tzeentch — return a random instant or sorcery", effects.ReturnRandomInstantOrSorceryFromYourGraveyard),
		},
	})
}
