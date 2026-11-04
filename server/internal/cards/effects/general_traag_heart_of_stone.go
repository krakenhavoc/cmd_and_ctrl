package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// General Traag, Heart of Stone — {3}{R}{R} Legendary Artifact Creature —
// Elemental Soldier 4/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 When General Traag enters, you may sacrifice another artifact. When
//	 you do, General Traag deals 4 damage to target creature."
//
// Bolg of the North's shape: the enters trigger asks "you may sacrifice
// another artifact" as it resolves (maySacrificeOne with another set, so
// Traag, an artifact himself, is never offered), and an artifact that really
// went raises the CR 603.12 reflexive trigger. That trigger chooses its
// target creature as it goes on the stack, after the sacrifice, and Traag
// deals the 4 damage (as he last existed, if he has left).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ade54bb7-3df8-4578-add8-1f85afee57ab",
		Name:            "General Traag, Heart of Stone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("General Traag — you may sacrifice another artifact", generalTraagMaySacrifice),
		},
	})
}

// generalTraagMaySacrifice is the enters trigger's body: the "may" prompt.
func generalTraagMaySacrifice(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "General Traag — sacrifice another artifact to deal 4 damage to target creature?",
		Artifact(), true, generalTraagSacrificed)
}

// generalTraagSacrificed raises the reflexive trigger once an artifact has
// really gone, keyed.
var generalTraagSacrificed = RegisterSacrificedOne("prompt-run/general-traag-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	t := WhenYouDo("General Traag — 4 damage to target creature", generalTraagStrikeBody)
	t.Params = game.EffectParams{Amount: 4}
	return t.Apply(ctx)
})

// generalTraagStrikeBody is the reflexive trigger: Params.Amount damage to
// the target creature chosen as it went on the stack.
var generalTraagStrikeBody = game.ReflexiveBody("general-traag-heart-of-stone/strike", breechesBlastEffect,
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
