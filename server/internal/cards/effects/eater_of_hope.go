package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eater of Hope — Creature — Demon {5}{B}{B}, 6/4:
//
//	"Flying
//	 {B}, Sacrifice another creature: Regenerate this creature.
//	 {2}{B}, Sacrifice two other creatures: Destroy target creature."
//
// Two sacrifice outlets. "Another" and "other" exclude the Demon itself
// by object (SacrificeAnotherN), so it can't pay for its own shield, and
// the two creatures of the second cost leave as one simultaneous exit.
// The shield is the engine's CR 701.19a regeneration shield: it replaces
// the next destruction this turn and is spent doing so.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3055381a-fb6b-47a3-8e77-a44768808e27",
		Name:            "Eater of Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{B}, Sacrifice another creature: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{B}"), SacrificeAnotherN(1, "another creature", Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{2}{B}, Sacrifice two other creatures: Destroy target creature.",
				Cost:    Plus(ManaCost("{2}{B}"), SacrificeAnotherN(2, "two other creatures", Creature())),
				Targets: TargetCreature("target creature"),
				Effect:  destroyFirstLegalCardTarget,
			},
		},
	})
}
