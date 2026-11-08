package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruthless Lawbringer — Creature — Vampire Assassin {1}{W}{B}, 3/2
// (Outlaws of Thunder Junction, #4468):
//
//	"When this creature enters, you may sacrifice another creature. When
//	 you do, destroy target nonland permanent."
//
// Gastal Blockbuster's shape with Bolg of the North's "another": the
// sacrifice is asked as the enters trigger resolves, over the other
// creatures you control then, with a floor of zero for the "may". Only
// once a creature has really been sacrificed is the destroy created, as a
// reflexive trigger that chooses its target as it goes on the stack (CR
// 603.12), so the target is picked after the sacrifice. Any nonland
// permanent is legal, the Lawbringer itself included; with none, the
// trigger is removed for want of a target (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82b6a283-9897-4ce1-8de3-e7579cc9b75b",
		Name:         "Ruthless Lawbringer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ruthless Lawbringer — you may sacrifice another creature", ruthlessLawbringerSacrifice),
		},
	})
}

func ruthlessLawbringerSacrifice(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "Ruthless Lawbringer — sacrifice another creature to destroy target nonland permanent?",
		Creature(), true, ruthlessLawbringerSacrificed)
}

// ruthlessLawbringerSacrificed is the "when you do". An on-disk key: never
// renamed, never reused.
var ruthlessLawbringerSacrificed = RegisterSacrificedOne("prompt-run/ruthless-lawbringer-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return WhenYouDo("Ruthless Lawbringer — destroy target nonland permanent", ruthlessLawbringerDestroyBody).Apply(ctx)
})

// ruthlessLawbringerDestroyBody is the reflexive trigger. An on-disk key:
// never renamed, never reused.
var ruthlessLawbringerDestroyBody = game.ReflexiveBody("ruthless-lawbringer/destroy", simpleBody(destroyFirstLegalCardTarget),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target nonland permanent", Nonland())
	}))
