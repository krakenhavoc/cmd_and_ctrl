package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selfcraft Mechan — Artifact Creature — Robot Artificer {3}{U}, 3/4 (Edge
// of Eternities, #4461):
//
//	"When this creature enters, you may sacrifice an artifact. When you
//	 do, put a +1/+1 counter on target creature and draw a card."
//
// Dire Blunderbuss's shape: the "you may" and the sacrifice belong to the
// enters trigger (the Mechan itself is an artifact and may be the one
// sacrificed), and the counter and the draw are a CR 603.12 reflexive
// trigger created only once an artifact has really gone. Its target is
// chosen as it goes on the stack (the ruling), and if the creature is an
// illegal target as it resolves, neither the counter nor the draw happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ba4cb4b-f421-4abd-8fcf-a01174447790",
		Name:         "Selfcraft Mechan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Selfcraft Mechan — you may sacrifice an artifact", selfcraftMechanEnters),
		},
	})
}

func selfcraftMechanEnters(g *game.Game, item *game.StackItem) error {
	if countControlled(g, item.Controller, func(c game.Card) bool { return c.IsArtifact() }) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Selfcraft Mechan — sacrifice an artifact?",
		YesLabel: "Sacrifice an artifact",
		NoLabel:  "Don't",
		Then:     selfcraftMechanAnswer,
	}.Apply(NewContext(g, item))
}

// selfcraftMechanAnswer is the "yes": sacrifice an artifact, then raise the
// reflexive trigger. Keyed: an on-disk identity, never renamed.
var selfcraftMechanAnswer = MayYes("confirm/selfcraft-mechan-sacrifice", func(ctx *Context, _ game.ConfirmAnswer) error {
	return ctx.SacrificeThen(ctx.Item.Controller,
		sacrificeSpec("an artifact", func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsArtifact() }),
		"Selfcraft Mechan — sacrifice an artifact",
		1, selfcraftMechanSacrificed)
})

// selfcraftMechanSacrificed raises the reflexive trigger once an artifact
// has really been sacrificed.
var selfcraftMechanSacrificed = RunThen("prompt-run/selfcraft-mechan-sacrificed", func(ctx *Context, r game.RunDone) error {
	if r.Sacrifices().Count() == 0 {
		return nil
	}
	return WhenYouDo("Selfcraft Mechan — put a +1/+1 counter on target creature and draw a card",
		selfcraftMechanCounterBody).Apply(ctx)
})

// selfcraftMechanCounterBody is the reflexive trigger: a +1/+1 counter on
// the target creature, then a draw. Neither happens if the target is gone.
var selfcraftMechanCounterBody = game.ReflexiveBody("selfcraft-mechan/counter-and-draw", simpleBody(selfcraftMechanCounter),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

func selfcraftMechanCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ResolveSteps(
			AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1},
			DrawCards{N: 1},
		)(item, ctx)
	}
	return nil
}
