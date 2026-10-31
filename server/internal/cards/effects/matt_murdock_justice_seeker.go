package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Matt Murdock, Justice Seeker — Legendary Creature — Human Advisor Hero
// {1}{W}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"At the beginning of combat on your turn, you may pay {1}. When you do,
//	 put a +1/+1 counter on target creature you control.
//	 Each creature you control with a counter on it has ward {1}."
//
// The first ability is Ant-Man, Colony Commander's MayPay / CR 603.12 pair
// (the target is chosen when the reflexive trigger goes on the stack, after
// the {1} is spent), but the counter goes on a creature you control. The
// second is Chapel Shieldgeist's WardGranted over the creatures its
// controller controls, narrowed to a creature with any counter on it as it
// becomes a target (a +1/+1 counter, a stun counter, any kind).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e64efd0d-3de3-42e2-949e-4a152d08def8",
		Name:         "Matt Murdock, Justice Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat(mattMurdockPayLabel, mattMurdockPayOffer),
			WardGranted(WardMana("{1}"), "Matt Murdock, Justice Seeker — each creature you control with a counter on it has ward {1}",
				func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller &&
						b41HasACounter()(g, source.Controller, *target)
				}),
		},
	})
}

const mattMurdockPayLabel = "Matt Murdock, Justice Seeker — pay {1} to put a +1/+1 counter on target creature you control"

// mattMurdockPayOffer asks for the {1}; the reflexive trigger lives in OnPay.
func mattMurdockPayOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: mattMurdockPayLabel + "?",
		Then:     mattMurdockPaid,
	}.Apply(NewContext(g, item))
}

// mattMurdockPaid is the paid half, keyed: the reflexive trigger is queued
// against the asking item.
var mattMurdockPaid = PayUnlessBranches("pay-unless/matt-murdock-justice-seeker-counter",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Matt Murdock, Justice Seeker — put a +1/+1 counter on target creature you control",
			Body:  mattMurdockCounterBody,
		}.Apply(ctx)
	}, nil)

// mattMurdockCounterBody is "when you do, put a +1/+1 counter on target
// creature you control".
var mattMurdockCounterBody = game.ReflexiveBody("matt-murdock-justice-seeker/counter", simpleBody(PlusOneCounterOnTheTarget),
	constTargets(func() *game.TargetSpec { return PermanentYouControl("target creature you control", Creature()) }))
