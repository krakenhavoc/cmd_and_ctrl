package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boom Box — Artifact {2} (Outlaws of Thunder Junction, #4468):
//
//	"{6}, {T}, Sacrifice this artifact: Destroy up to one target
//	 artifact, up to one target creature, and up to one target land."
//
// Giant's Boulder's cost, with Teferi, Who Slows the Sunset's three "up to
// one" clauses (Clauses + WithCount(0, 1)) and Urgent Necropsy's
// resolution: the targets still legal are destroyed in one simultaneous
// event (destroyEachLegalTarget), so an artifact creature chosen for two
// of the words is destroyed once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0fdf2e0-3025-4919-a9dc-d33100e487fd",
		Name:         "Boom Box",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{6}, {T}, Sacrifice this artifact: Destroy up to one target artifact, up to one target creature, and up to one target land.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{6}"},
			Targets: Clauses(
				TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1),
				TargetCreature("up to one target creature").WithCount(0, 1),
				TargetPermanent("up to one target land", Land()).WithCount(0, 1),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return destroyEachLegalTarget(item, NewContext(g, item))
			},
		}},
	})
}
