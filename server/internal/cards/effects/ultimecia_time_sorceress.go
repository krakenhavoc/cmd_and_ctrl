package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ultimeciaOracle is shared by both faces: the front registers under it,
// the back under ultimeciaOracle + "#1".
const ultimeciaOracle = "70773879-b773-4e81-99ca-8f4bba478962"

// Ultimecia, Time Sorceress // Ultimecia, Omnipotent — a transforming
// legend (Final Fantasy, #4462):
//
//	Ultimecia, Time Sorceress, {3}{U}{B} Legendary Creature — Human
//	Warlock, 4/5
//	  "Whenever Ultimecia enters or attacks, surveil 2.
//	   At the beginning of your end step, you may pay {4}{U}{U}{B}{B} and
//	   exile eight cards from your graveyard. If you do, transform
//	   Ultimecia."
//
//	Ultimecia, Omnipotent, Legendary Creature — Nightmare Warlock, 7/7
//	  "Menace
//	   Time Compression — When this creature transforms into Ultimecia,
//	   Omnipotent, take an extra turn after this one."
//
// The surveil is one trigger for two conditions (enters or attacks) over a
// surveil step. The end-step ability is Banon's pay-then-act shape with
// the second half of the cost being the exile: a MayPay of {4}{U}{U}{B}{B}
// (an unfundable "yes" is the decline) whose paid branch opens a pick of
// exactly eight cards from your graveyard, exiled together, and only then
// does Ultimecia transform (TransformThis, which does nothing if it has
// turned over already, CR 701.27f). Both halves of the cost are owed once
// the player says yes; with fewer than eight cards in the graveyard the
// cost cannot be paid in full (CR 118.3), so nothing is offered rather
// than taking the mana for nothing. Time Compression triggers on the
// transform (CR 701.27e, ThisTransformedInto) and is a TakeExtraTurn; a
// permanent that enters on the back face never transformed, so takes none
// (CR 712.18).
//
// No simplification.
func init() {
	const surveilLabel = "Ultimecia, Time Sorceress — surveil 2"
	Register(Spec{
		OracleID:     ultimeciaOracle,
		Name:         "Ultimecia, Time Sorceress",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks(surveilLabel, Do(Surveil{N: 2})),
			AtYourEndStep("Ultimecia, Time Sorceress — you may pay {4}{U}{U}{B}{B} and exile eight cards from your graveyard; if you do, transform Ultimecia",
				func(g *game.Game, item *game.StackItem) error {
					if len(graveyardCardsOfPlayer(g, item.Controller)) < 8 {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{4}{U}{U}{B}{B}",
						Question: "Ultimecia — pay {4}{U}{U}{B}{B} and exile eight cards from your graveyard to transform Ultimecia?",
						Then:     ultimeciaPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:        ultimeciaOracle + "#1",
		Name:            "Ultimecia, Omnipotent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventTransform, ThisTransformedInto(1),
				"Ultimecia, Omnipotent — Time Compression: take an extra turn after this one",
				Do(TakeExtraTurn{})), game.Purpose{ExtraTurns: 1}),
		},
	})
}

// ultimeciaPaid is the paid half of the end-step trigger, keyed (#4264):
// the mana is paid, so the pick of eight cards is owed. It is a keyed
// choose-cards prompt (ultimeciaExile), so a table waiting on it is a
// restore point.
var ultimeciaPaid = PayUnlessBranches("pay-unless/ultimecia-exile",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		cards := graveyardCardsOfPlayer(ctx.Game, ctx.Controller())
		if len(cards) < 8 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Ultimecia — exile eight cards from your graveyard, then transform Ultimecia",
			Cards:    cards,
			Min:      8,
			Max:      8,
			ThenKey:  ultimeciaExile,
			LiveZone: game.ZoneGraveyard,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}, nil)

// ultimeciaExile exiles the eight cards picked, then transforms Ultimecia
// if it is still the permanent that asked. An on-disk key.
var ultimeciaExile = askedPick("choose-cards/ultimecia-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 8 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(ultimeciaTransform))
})

// ultimeciaTransform is "If you do, transform Ultimecia", once the cards
// are exiled. An on-disk key.
var ultimeciaTransform = CardTail("ultimecia/transform", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) < 8 {
		return nil
	}
	return TransformThis{}.Apply(ctx)
})
