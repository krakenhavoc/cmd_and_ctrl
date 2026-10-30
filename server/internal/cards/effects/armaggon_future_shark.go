package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armaggon, Future Shark — {6}{B}{B} Legendary Creature — Shark Horror
// Mutant 9/6 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flash
//	 When Armaggon enters, destroy up to three target creatures."
//
// Flash is the printed keyword. The enters trigger is Finale of Eternity's
// "destroy up to three target creatures" without the toughness bound: zero
// to three targets, any controller, destroyed together in one event so a
// death trigger among them sees the others (destroyEachLegalTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e2a35850-92c5-40ab-8697-2e995819d7ac",
		Name:            "Armaggon, Future Shark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Armaggon — destroy up to three target creatures",
					func(g *game.Game, item *game.StackItem) error {
						return destroyEachLegalTarget(item, NewContext(g, item))
					}),
				TargetCreature("up to three target creatures").WithCount(0, 3)),
		},
	})
}
