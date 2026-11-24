package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Null Caller — Creature — Vampire Shaman {3}{B}, 2/4 (Commander Legends):
//
//	"{3}{B}, Exile a creature card from your graveyard: Create a tapped 2/2
//	 black Zombie creature token."
//
// Moorland Haunt's cost (a creature card from the controller's own
// graveyard, exiled as the cost is paid, with no {T}, so summoning sickness
// does not apply) over Altar of Bhaal's tapped token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9d75969-15ac-4509-b95c-a762fc2b99f9",
		Name:         "Null Caller",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, Exile a creature card from your graveyard: Create a tapped 2/2 black Zombie creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateTokenAdvanced{
					Spec: effects.Token(effects.TokenCard("2/2 black Zombie")).EntersTapped(),
					N:    1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
