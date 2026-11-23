package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefield Butcher — Creature — Human Soldier {2}{B}, 1/4 (The Brothers' War):
//
//	"{5}, {T}: Each opponent loses 2 life. This ability costs {1} less to
//	 activate for each creature card in your graveyard."
//
// Diary of Dreams's ability cost modifier (CR 602.2b, 601.2f) counted with
// Cyan, Vengeful Samurai's creature-card count: only the generic {5} is
// reduced, and the {T} is still paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2978ba77-c0ce-4411-99c7-ef05dd602c8d",
		Name:         "Battlefield Butcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Each opponent loses 2 life. This ability costs {1} less to activate for each creature card in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.CreatureCardsInYourGraveyard,
					"This ability costs {1} less to activate for each creature card in your graveyard."),
			},
			Effect: effects.Do(effects.EachOpponentLosesLife{N: 2}),
		}},
	})
}
