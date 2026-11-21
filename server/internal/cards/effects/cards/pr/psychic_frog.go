package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Frog — Creature — Frog {U}{B}, 1/2 (Modern Horizons 3):
//
//	"Whenever this creature deals combat damage to a player or planeswalker,
//	 draw a card.
//	 Discard a card: Put a +1/+1 counter on this creature.
//	 Exile three cards from your graveyard: This creature gains flying until end
//	 of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e157ad1e-be35-47f3-92db-01799f8fb6a5",
		Name:         "Psychic Frog",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Psychic Frog — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Discard a card: Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.DiscardACard(),
				Effect:  effects.PutCounterOnSelf,
			},
			{
				Label:   "Exile three cards from your graveyard: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ExileFromGraveyard(3, "three cards from your graveyard", nil),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
						Label: "Psychic Frog — flying until end of turn"}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
