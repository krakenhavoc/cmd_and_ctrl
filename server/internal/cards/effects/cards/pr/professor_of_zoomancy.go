package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Professor of Zoomancy — Creature — Bear Druid {3}{G}, 4/3 (Strixhaven:
// School of Mages, #5502):
//
//	"When this creature enters, create a 1/1 black and green Pest creature
//	 token with "When this token dies, you gain 1 life.""
//
// The Strixhaven Pest (PestToken, with its own dies trigger), made by an
// enters trigger on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a22b8220-b41d-4b58-aa68-2c65ce37c517",
		Name:         "Professor of Zoomancy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Professor of Zoomancy — create a 1/1 Pest that gains you 1 life when it dies",
				effects.Do(effects.CreateToken{Template: effects.PestToken(), N: 1})),
		},
	})
}
