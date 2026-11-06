package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cornered Crook — Creature — Lizard Warrior {4}{R}, 5/4 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature enters, you may sacrifice an artifact. When you
//	 do, this creature deals 3 damage to any target."
//
// General Traag, Heart of Stone's shape: the enters trigger asks "you
// may sacrifice an artifact" as it resolves (maySacrificeOne), and an
// artifact that really went raises the CR 603.12 reflexive trigger,
// which chooses its target as it goes on the stack, after the
// sacrifice. The Crook deals the 3 damage, as it last existed if it
// has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4477eeab-a51a-4dc2-9c97-83f2551d8b54",
		Name:         "Cornered Crook",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cornered Crook — you may sacrifice an artifact",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Cornered Crook — sacrifice an artifact to deal 3 damage to any target?",
						Artifact(), false, corneredCrookSacrificed)
				}),
		},
	})
}

// corneredCrookSacrificed raises the reflexive trigger once an artifact
// has really gone, keyed.
var corneredCrookSacrificed = RegisterSacrificedOne("prompt-run/cornered-crook-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	t := WhenYouDo("Cornered Crook — 3 damage to any target", corneredCrookDamageBody)
	t.Params = game.EffectParams{Amount: 3}
	return t.Apply(ctx)
})

// corneredCrookDamageBody is the reflexive trigger: Params.Amount
// damage to the target chosen as it went on the stack.
var corneredCrookDamageBody = game.ReflexiveBody("cornered-crook/damage", breechesBlastEffect, constTargets(TargetAny))
