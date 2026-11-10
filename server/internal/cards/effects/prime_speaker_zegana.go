package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prime Speaker Zegana — Legendary Creature — Merfolk Wizard
// {2}{G}{G}{U}{U}, 1/1 (Foundations, #4465):
//
//	"Prime Speaker Zegana enters with X +1/+1 counters on it, where X is
//	 the greatest power among other creatures you control.
//	 When Prime Speaker Zegana enters, draw cards equal to its power."
//
// The counters are a CR 614.1c entry clause whose X is counted once, as
// Zegana enters (b19EntersWithCountersCounted, Diregraf Colossus's), while
// it is still the stack object: it is never "another creature", and a
// creature entering at the same time isn't counted (the ruling). With no other
// creature X is 0. The draw reads Zegana's power as it resolves, counters
// included, and from last-known information if she has left (CR 608.2h,
// the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "311e9368-696a-47e7-aa2f-3ef1b1a92e2b",
		Name:         "Prime Speaker Zegana",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return b42GreatestPowerControlledBy(g, src.Controller)
			}, "Prime Speaker Zegana: enters with X +1/+1 counters, X the greatest power among other creatures you control"),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisEnters("Prime Speaker Zegana — draw cards equal to its power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || info.Power <= 0 {
						return nil
					}
					return DrawCards{Player: item.Controller, N: info.Power}.Apply(ctx)
				}), game.Purpose{}),
		},
	})
}
