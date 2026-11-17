package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Terror Ballista — Artifact Creature — Construct {7}, 5/3 (The Brothers'
// War, #5482):
//
//	"Menace
//	 Whenever this creature attacks, you may sacrifice another creature. When
//	 you do, destroy target creature an opponent controls.
//	 Unearth {3}{B}{B} (…)"
//
// Gastal Blockbuster's shape: the sacrifice is asked as the attack trigger
// resolves, over the OTHER creatures you control then, with a floor of zero
// for the "may". Only a creature that really left creates the reflexive
// trigger, which targets as it goes on the stack (CR 603.12) and is removed
// for want of a target (CR 603.3d). Unearth is the shared keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3044a00d-7f88-47e8-9110-2e898b20416c",
		Name:            "Terror Ballista",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Terror Ballista — you may sacrifice another creature to destroy target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Terror Ballista — sacrifice another creature to destroy target creature an opponent controls?",
						Creature(), true, terrorBallistaSacrificed)
				}),
		},
		Activated: []ActivatedAbility{Unearth("{3}{B}{B}")},
	})
}

// terrorBallistaSacrificed is the "when you do". An on-disk key: never
// renamed, never reused.
var terrorBallistaSacrificed = RegisterSacrificedOne("prompt-run/terror-ballista-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return WhenYouDo("Terror Ballista — destroy target creature an opponent controls", terrorBallistaDestroyBody).Apply(ctx)
})

// terrorBallistaDestroyBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var terrorBallistaDestroyBody = game.ReflexiveBody("terror-ballista/destroy", simpleBody(destroyFirstLegalCardTarget),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target creature an opponent controls", Creature(), OpponentControls())
	}))
