package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunstar Chaplain — Creature — Human Cleric {1}{W}, 3/2 (Edge of
// Eternities, #4461):
//
//	"At the beginning of your end step, if you control two or more tapped
//	 creatures, put a +1/+1 counter on target creature you control.
//	 {2}, Remove a +1/+1 counter from a creature you control: Tap target
//	 artifact or creature."
//
// The end-step trigger is Dawnstrike Vanguard's intervening "if"
// (CR 603.4), checked as the step begins and again as it resolves, with
// one target creature you control chosen as it goes on the stack. The
// activation is Ray Fillet's "remove a +1/+1 counter from a creature you
// control" cost (the creature is chosen as the cost is paid, not
// targeted) with Swashbuckler's Whip's tap. No {T}, so it can be
// activated the turn the Chaplain arrives, as often as it can be paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de6b26e4-8700-490f-9df2-dac430463274",
		Name:         "Sunstar Chaplain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(InterveningIf(CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *Context) bool { return b24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2 },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Sunstar Chaplain — put a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget)),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Remove a +1/+1 counter from a creature you control: Tap target artifact or creature.",
			Cost:    Plus(ManaCost("{2}"), RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", Creature())),
			Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
			Effect:  tapTheFirstLegalTarget,
		}},
	})
}
