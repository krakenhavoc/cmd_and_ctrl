package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subspace Anomaly — Land (Star Trek, #4973):
//
//	"{T}: Add {C}.
//	 {4}{W}, {T}, Sacrifice this land: Create two 1/1 red Officer creature
//	 tokens. Create a tapped Planet land token named New Planet with
//	 "{T}: Add one mana of any color.""
//
// The Officers are the printed token (ttrk 6); the New Planet is
// Genesis Effect's (ttrk 27). The sacrifice is a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a28a83a-de27-4aaa-b35f-1e7075fc1e43",
		Name:         "Subspace Anomaly",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{4}{W}, {T}, Sacrifice this land: Create two 1/1 red Officer creature tokens. Create a tapped Planet land token named New Planet.",
			Cost:  effects.TrekB29AnomalyCost("{4}{W}"),
			// Two Officers that can block, and a land.
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 3, CreatureTokens: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.CreateToken{Template: effects.TokenCard("1/1 red Officer"), N: 2}).Apply(ctx); err != nil {
					return err
				}
				return effects.TrekB29NewPlanet(ctx)
			},
		}},
	})
}
