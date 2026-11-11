package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dream Eater — Creature — Nightmare Sphinx {4}{U}{U}, 4/3:
//
//	"Flash
//	 Flying
//	 When this creature enters, surveil 4. When you do, you may return
//	 target nonland permanent an opponent controls to its owner's hand."
//
// The surveil is a keyed prompt, and the rest of the sentence is its tail
// (CR 603.12): once the cards are placed, a reflexive "when you do" trigger
// goes on the stack, targeting a nonland permanent an opponent controls, and
// the "may" is asked as it resolves. A surveil with an empty library looks at
// nothing and does not happen, so no reflexive trigger follows it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4ec84bb3-796a-4074-ac9a-1bd277158896",
		Name:            "Dream Eater",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dream Eater — surveil 4; when you do, you may return target nonland permanent an opponent controls to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return Surveil{Player: ctx.Controller(), N: 4, ThenTail: ctx.ItemTail(dreamEaterThen)}.Apply(ctx)
				}),
		},
	})
}

// dreamEaterThen is "When you do, you may return target nonland permanent an
// opponent controls to its owner's hand", once the surveil is done. An
// on-disk key: never renamed, never reused.
var dreamEaterThen = CardTail("then/dream-eater-reflexive-bounce", func(ctx *Context, _ game.TailOutcome) error {
	r := WhenYouDo("Dream Eater — you may return target nonland permanent an opponent controls to its owner's hand", dreamEaterBounceBody)
	r.Optional = "Return target nonland permanent an opponent controls to its owner's hand?"
	return r.Apply(ctx)
})

var dreamEaterBounceBody = game.ReflexiveBody("dream-eater/bounce",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return BounceToHand{Target: t.ID}.Apply(ctx)
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls())
	}))
