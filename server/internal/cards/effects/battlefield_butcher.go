package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2978ba77-c0ce-4411-99c7-ef05dd602c8d",
		Name:         "Battlefield Butcher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Each opponent loses 2 life. This ability costs {1} less to activate for each creature card in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(creatureCardsInYourGraveyard,
					"This ability costs {1} less to activate for each creature card in your graveyard."),
			},
			Effect: Do(EachOpponentLosesLife{N: 2}),
		}},
	})
}
