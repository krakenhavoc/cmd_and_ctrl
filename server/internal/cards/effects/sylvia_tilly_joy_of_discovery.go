package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvia Tilly, Joy of Discovery — {2}{G} Legendary Creature — Human
// Scientist Artificer 2/2 (Star Trek, #4973):
//
//	"When Sylvia Tilly enters, surveil 2.
//	 At the beginning of combat on your turn, you may pay {1} and exile a
//	 card from your graveyard. When you do, put a +1/+1 counter on target
//	 creature."
//
// The combat trigger is Banon's and Conspiracy Theorist's "pay {1} and
// discard a card" shape with a graveyard in place of the hand: the offer
// is made only while there is a card in your graveyard to exile, the {1}
// is paid, and then you choose the card. Only once a card has really been
// exiled does the reflexive trigger go on the stack and take its target
// (CR 603.12); with no creature to target it is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4c1dbc8-23d8-49e0-bc2d-d005e9215172",
		Name:         "Sylvia Tilly, Joy of Discovery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sylvia Tilly — surveil 2", Do(Surveil{N: 2})),
			AtBeginningOfYourCombat(sylviaTillyCombatLabel, sylviaTillyOffer),
		},
	})
}

const sylviaTillyCombatLabel = "Sylvia Tilly — pay {1} and exile a card from your graveyard to put a +1/+1 counter on target creature"

func sylviaTillyOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(graveyardCardIDs(ctx, item.Controller, trekB22AnyCard)) == 0 {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Sylvia Tilly — pay {1} and exile a card from your graveyard to put a +1/+1 counter on target creature?",
		Then:     sylviaTillyPaid,
	}.Apply(ctx)
}

var sylviaTillyPaid = PayUnlessBranches("pay-unless/sylvia-tilly-joy-of-discovery-exile",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		controller := ctx.Controller()
		candidates := graveyardCardIDs(ctx, controller, trekB22AnyCard)
		if len(candidates) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    controller,
			FromPlayer: controller,
			Source:     ctx.Source(),
			Question:   "Sylvia Tilly — exile a card from your graveyard",
			Cards:      candidates,
			Min:        1,
			Max:        1,
			ThenKey:    sylviaTillyExiled,
			LiveZone:   game.ZoneGraveyard,
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}, nil)

var sylviaTillyExiled = askedPick("choose-cards/sylvia-tilly-joy-of-discovery-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(sylviaTillyWhenYouDo)}.Apply(ctx)
})

// sylviaTillyWhenYouDo is "When you do, put a +1/+1 counter on target
// creature", run once the exile has really happened: a card a CR 614
// replacement sent elsewhere is not "exiled", and no reflexive trigger
// goes on the stack.
var sylviaTillyWhenYouDo = CardTail("then/sylvia-tilly-joy-of-discovery-when-you-do", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return ReflexiveTrigger{
		Label: "Sylvia Tilly — put a +1/+1 counter on target creature",
		Body:  sylviaTillyCounterBody,
	}.Apply(ctx)
})

var sylviaTillyCounterBody = game.ReflexiveBody("sylvia-tilly-joy-of-discovery/counter", simpleBody(PlusOneCounterOnTheTarget),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
