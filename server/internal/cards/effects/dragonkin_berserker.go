package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonkin Berserker — Creature — Human Berserker {1}{R}, 2/2:
//
//	"First strike
//	 Boast abilities you activate cost {1} less to activate for each Dragon you control.
//	 Boast — {4}{R}: Create a 5/5 red Dragon creature token with flying. (Activate only if
//	 this creature attacked this turn and only once each turn.)"
//
// The discount is a board cost modifier scoped to activations
// (CostModifier.Activations), so it reaches every boast ability its
// controller activates, this card's own included, from any creature.
// Only the generic part of a cost can be reduced (CR 601.2f), exactly
// as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06b77a80-4968-4385-960a-fb66bb4faa94",
		Name:            "Dragonkin Berserker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		CostModifiers: []game.CostModifier{
			ActivationCostsLessEach(PermanentsYouControl(And(Creature(), OfCreatureType("Dragon"))),
				"Boast abilities you activate cost {1} less to activate for each Dragon you control.",
				ABoastAbilityCost(), ActivatedByTheModifiersController()),
		},
		Activated: []ActivatedAbility{
			BoastAnswering(game.AnswerMakesBlocker, "{4}{R}: Create a 5/5 red Dragon creature token with flying.",
				ManaCost("{4}{R}"), createTheToken("5/5 red Dragon with flying")),
		},
	})
}
