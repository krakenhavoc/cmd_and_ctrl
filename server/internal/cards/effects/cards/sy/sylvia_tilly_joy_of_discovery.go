package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d4c1dbc8-23d8-49e0-bc2d-d005e9215172",
		Name:         "Sylvia Tilly, Joy of Discovery",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sylvia Tilly — surveil 2", effects.Do(effects.Surveil{N: 2})),
			effects.AtBeginningOfYourCombat(sylviaTillyCombatLabel, sylviaTillyOffer),
		},
	})
}

const sylviaTillyCombatLabel = "Sylvia Tilly — pay {1} and exile a card from your graveyard to put a +1/+1 counter on target creature"

func sylviaTillyOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(effects.GraveyardCardIDs(ctx, item.Controller, effects.TrekB22AnyCard)) == 0 {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Sylvia Tilly — pay {1} and exile a card from your graveyard to put a +1/+1 counter on target creature?",
		Then:     sylviaTillyPaid,
	}.Apply(ctx)
}

var sylviaTillyPaid = effects.PayUnlessBranches("pay-unless/sylvia-tilly-joy-of-discovery-exile",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		controller := ctx.Controller()
		candidates := effects.GraveyardCardIDs(ctx, controller, effects.TrekB22AnyCard)
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

var sylviaTillyExiled = effects.AskedPick("choose-cards/sylvia-tilly-joy-of-discovery-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(sylviaTillyWhenYouDo)}.Apply(ctx)
})

// sylviaTillyWhenYouDo is "When you do, put a +1/+1 counter on target
// creature", run once the exile has really happened: a card a CR 614
// replacement sent elsewhere is not "exiled", and no reflexive trigger
// goes on the stack.
var sylviaTillyWhenYouDo = effects.CardTail("then/sylvia-tilly-joy-of-discovery-when-you-do", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label: "Sylvia Tilly — put a +1/+1 counter on target creature",
		Body:  sylviaTillyCounterBody,
	}.Apply(ctx)
})

var sylviaTillyCounterBody = game.ReflexiveBody("sylvia-tilly-joy-of-discovery/counter", effects.SimpleBody(effects.PlusOneCounterOnTheTarget),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))
