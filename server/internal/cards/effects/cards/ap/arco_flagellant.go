package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arco-Flagellant — Creature — Human {2}{B}, 3/1 (Warhammer 40,000
// Commander, S82 GAP batch 187, #5948):
//
//	"Squad {2}
//	 This creature can't block.
//	 Endurant — Pay 3 life: This creature gains indestructible until end of
//	 turn."
//
// Squad is the shared keyword (CR 702.157); the life is a cost, so a player
// with fewer than three life can't activate it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b48e7e87-3adf-417d-b93b-bf67ad6cbf1d",
		Name:          "Arco-Flagellant",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		Static:        []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered:     []game.TriggeredAbility{effects.Squad("Arco-Flagellant")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 3 life: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.PayLife(3),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Arco-Flagellant — indestructible until end of turn"),
		}},
	})
}
