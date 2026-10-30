package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Thing — Legendary Creature — Human Hero {5}{G}, 5/5 (Marvel Super
// Heroes Commander, #4455):
//
//	"Trample
//	 At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, put four +1/+1 counters on The Thing.
//	 Whenever The Thing attacks, you may pay {R}{G}{W}{U}. When you do,
//	 double the number of each kind of counter on any number of target
//	 permanents you control."
//
// The counters are an intervening-if trigger on the per-turn cast tally
// (Mister Fantastic's). The attack trigger is Ant-Man, Colony Commander's
// MayPay / CR 603.12 pair: the payment is the whole condition and the
// doubling is a second triggered ability whose targets (any number of
// permanents you control, none allowed) are chosen as it goes on the stack.
// Each kind of counter on each target is read once and added again through
// AddCounter (Vorel of the Hull Clade's doubling), so a replacement such as
// Hardened Scales applies, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8de07d4-9bc1-4bfe-b5a4-0a58e05af022",
		Name:            "The Thing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			mscM13CombatIfCastNoncreature("The Thing — put four +1/+1 counters on The Thing",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 4})),
			WheneverThisAttacks(theThingPayLabel, theThingPayOffer),
		},
	})
}

const theThingPayLabel = "The Thing — pay {R}{G}{W}{U} to double the counters on any number of target permanents you control"

// theThingPayOffer asks for {R}{G}{W}{U}; the reflexive trigger lives in OnPay.
func theThingPayOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{R}{G}{W}{U}",
		Question: theThingPayLabel + "?",
		Then:     theThingPaid,
	}.Apply(NewContext(g, item))
}

// theThingPaid is the paid half, keyed: the reflexive trigger is queued
// against the asking item.
var theThingPaid = PayUnlessBranches("pay-unless/the-thing-double-counters",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "The Thing — double the number of each kind of counter on any number of target permanents you control",
			Body:  theThingDoubleBody,
		}.Apply(ctx)
	}, nil)

// theThingDoubleBody is "when you do, double the number of each kind of
// counter on any number of target permanents you control".
var theThingDoubleBody = game.ReflexiveBody("the-thing/double-counters", simpleBody(theThingDouble),
	constTargets(func() *game.TargetSpec {
		return PermanentYouControl("any number of target permanents you control").WithCount(0, 0)
	}))

func theThingDouble(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := b30DoubleEachCounterKindOn(ctx, t.ID); err != nil {
			return err
		}
	}
	return nil
}
