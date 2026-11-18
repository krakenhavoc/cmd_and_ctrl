package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a949c96c-362c-45a3-bd5c-ce5ce153ee9e",
		Name:         "Terrain Generator",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: You may put a basic land card from your hand onto the battlefield tapped.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: Do(PutFromHandOntoBattlefield{
				Match:    rprBasicLandCard(),
				Optional: true,
				Tapped:   true,
				Label:    "Terrain Generator — you may put a basic land card from your hand onto the battlefield tapped",
			}),
		}},
	})
}
