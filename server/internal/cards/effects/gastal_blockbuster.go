package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gastal Blockbuster — Creature — Human Berserker {2}{R}, 3/2
// (Aetherdrift, #4464):
//
//	"When this creature enters, you may sacrifice a creature or Vehicle.
//	 When you do, destroy target artifact an opponent controls."
//
// Bolg of the North's shape: the sacrifice is asked as the trigger
// resolves, over the creatures and Vehicles you control then (the
// Blockbuster itself among them, since the text doesn't say "another"),
// with a floor of zero for the "may". Only once something has really
// been sacrificed is the destroy created, as a reflexive trigger that
// targets as it goes on the stack (CR 603.12). With no artifact an
// opponent controls, it is removed for want of a target (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b601721-90ca-45d3-9a3a-19c719799bfa",
		Name:         "Gastal Blockbuster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gastal Blockbuster — you may sacrifice a creature or Vehicle", gastalBlockbusterSacrifice),
		},
	})
}

func gastalBlockbusterSacrifice(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "Gastal Blockbuster — sacrifice a creature or Vehicle to destroy target artifact an opponent controls?",
		Or(Creature(), HasSubtype("Vehicle")), false, gastalBlockbusterSacrificed)
}

// gastalBlockbusterSacrificed is the "when you do". An on-disk key: never
// renamed, never reused.
var gastalBlockbusterSacrificed = RegisterSacrificedOne("prompt-run/gastal-blockbuster-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return WhenYouDo("Gastal Blockbuster — destroy target artifact an opponent controls", gastalBlockbusterDestroyBody).Apply(ctx)
})

// gastalBlockbusterDestroyBody is the reflexive trigger. An on-disk key:
// never renamed, never reused.
var gastalBlockbusterDestroyBody = game.ReflexiveBody("gastal-blockbuster/destroy", simpleBody(destroyFirstLegalCardTarget),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target artifact an opponent controls", Artifact(), OpponentControls())
	}))
