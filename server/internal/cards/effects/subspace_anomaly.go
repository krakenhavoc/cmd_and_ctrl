package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1a28a83a-de27-4aaa-b35f-1e7075fc1e43",
		Name:         "Subspace Anomaly",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{4}{W}, {T}, Sacrifice this land: Create two 1/1 red Officer creature tokens. Create a tapped Planet land token named New Planet.",
			Cost:  trekB29AnomalyCost("{4}{W}"),
			// Two Officers that can block, and a land.
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 3, CreatureTokens: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CreateToken{Template: TokenCard("1/1 red Officer"), N: 2}).Apply(ctx); err != nil {
					return err
				}
				return trekB29NewPlanet(ctx)
			},
		}},
	})
}
