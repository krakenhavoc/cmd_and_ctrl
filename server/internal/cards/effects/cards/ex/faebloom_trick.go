package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faebloom Trick — Instant {2}{U} (Foundations, #4465):
//
//	"Create two 1/1 blue Faerie creature tokens with flying. When you do,
//	 tap target creature an opponent controls."
//
// Ant-Man's reflexive shape (CR 603.12): the spell makes the two Faeries as
// it resolves and that creates a trigger, which goes on the stack above it
// and chooses its target there (CR 603.3d). With no opposing creature to
// tap the trigger is removed and the Faeries stay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "903d114b-1899-4e67-bee3-af0673850388",
		Name:         "Faebloom Trick",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 blue Faerie with flying"), N: 2},
			effects.ReflexiveTrigger{
				Label: "Faebloom Trick — tap target creature an opponent controls",
				Body:  faebloomTrickTapBody,
			},
		),
	})
}

// faebloomTrickTapBody is "when you do, tap target creature an opponent
// controls". An on-disk key.
var faebloomTrickTapBody = game.ReflexiveBody("faebloom-trick/tap", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.TapTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}), effects.ConstTargets(func() *game.TargetSpec {
	return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
}))
