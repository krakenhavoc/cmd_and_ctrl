package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terrain Generator — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: You may put a basic land card from your hand onto the
//	 battlefield tapped."
//
// Arboreal Grazer's put-from-hand clause narrowed to basic land cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a949c96c-362c-45a3-bd5c-ce5ce153ee9e",
		Name:         "Terrain Generator",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: You may put a basic land card from your hand onto the battlefield tapped.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: effects.Do(effects.PutFromHandOntoBattlefield{
				Match:    effects.RprBasicLandCard(),
				Optional: true,
				Tapped:   true,
				Label:    "Terrain Generator — you may put a basic land card from your hand onto the battlefield tapped",
			}),
		}},
	})
}
