package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nurturing Pixie — Creature — Faerie Rogue {W}, 1/1 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flying
//	 When this creature enters, return up to one target non-Faerie,
//	 nonland permanent you control to its owner's hand. If a permanent
//	 was returned this way, put a +1/+1 counter on this creature."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), and
// the Pixie, a Faerie, can never be its own target. The return is
// BounceToHand with a keyed tail, so "if a permanent was returned this
// way" reads where the permanent went (Bob, Reluctant HYDRA Agent's
// oneWentThisWay): no target, a target that is no longer legal, or a
// commander its owner sent to the command zone instead puts no counter
// on the Pixie.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ecd3743-7c32-41ce-9ac1-c8c855a47302",
		Name:            "Nurturing Pixie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Nurturing Pixie — return up to one target non-Faerie, nonland permanent you control to its owner's hand; if you do, put a +1/+1 counter on this creature",
				nurturingPixieReturn),
				TargetPermanent("up to one target non-Faerie, nonland permanent you control",
					Nonland(), Not(HasSubtype("Faerie")), YouControl()).WithCount(0, 1)),
		},
	})
}

// nurturingPixieReturn is the trigger's resolution: the bounce, and the
// counter hung off it.
func nurturingPixieReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return BounceToHand{Target: t.ID, ThenTail: ctx.ItemTail(nurturingPixieCounter)}.Apply(ctx)
	}
	return nil
}

// nurturingPixieCounter is "if a permanent was returned this way, put a
// +1/+1 counter on this creature".
var nurturingPixieCounter = CardTail("nurturing-pixie/counter-if-returned", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
