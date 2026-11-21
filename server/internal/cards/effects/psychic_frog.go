package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psychic Frog — Creature — Frog {U}{B}, 1/2 (Modern Horizons 3):
//
//	"Whenever this creature deals combat damage to a player or planeswalker,
//	 draw a card.
//	 Discard a card: Put a +1/+1 counter on this creature.
//	 Exile three cards from your graveyard: This creature gains flying until end
//	 of turn."
func init() {
	Register(Spec{
		OracleID:     "e157ad1e-be35-47f3-92db-01799f8fb6a5",
		Name:         "Psychic Frog",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Psychic Frog — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Discard a card: Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    DiscardACard(),
				Effect:  putCounterOnSelf,
			},
			{
				Label:   "Exile three cards from your graveyard: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ExileFromGraveyard(3, "three cards from your graveyard", nil),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
						Label: "Psychic Frog — flying until end of turn"}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
