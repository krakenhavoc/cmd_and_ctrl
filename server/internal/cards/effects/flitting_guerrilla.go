package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flitting Guerrilla — Creature — Faerie Rogue {2}{B}, 2/2 (March of the
// Machine):
//
//	"Flying
//	 When this creature dies, each player mills two cards. Then you may
//	 exile this card. When you do, put target creature or battle card
//	 from your graveyard on top of your library."
//
// The mills run in APNAP order, then the question: "this card" is the
// Guerrilla in the graveyard, and only if it is still there is it offered
// (it may have been milled-over or moved by a response). Exiling it is a
// reflexive trigger (CR 603.12) that goes on the stack with its target
// chosen then, after the mills, so a creature or battle card the mill just
// put into your graveyard is a legal target. The target is chosen from
// your graveyard as the reflexive trigger goes on the stack, and the
// Guerrilla, now in exile, is not in it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "24dbddad-998b-4755-b356-4c8aca3592b1",
		Name:            "Flitting Guerrilla",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Flitting Guerrilla — each player mills two cards, then you may exile this card; when you do, put target creature or battle card from your graveyard on top of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return millEachPlayerInTurn(ctx, apnapPlayers(g), constantMill(2), ctx.ItemTail(s82MomB12GuerrillaAfterMill))
				}),
		},
	})
}

// s82MomB12GuerrillaTopBody is the reflexive "put target creature or battle
// card from your graveyard on top of your library". An on-disk key.
var s82MomB12GuerrillaTopBody = game.ReflexiveBody("s82-mom-b12/flitting-guerrilla-top", simpleBody(PutChosenTargetOnTopOfLibrary),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature or battle card from your graveyard", Or(Creature(), Battle()), YouOwn())
	}))

// s82MomB12GuerrillaAfterMill is "then you may exile this card", run once
// every mill is done: asked only while the Guerrilla is still in a
// graveyard. An on-disk key.
var s82MomB12GuerrillaAfterMill = CardTail("tail/s82-mom-b12-guerrilla-after-mill", func(ctx *Context, _ game.TailOutcome) error {
	z := ctx.Game.FindCardZoneForEffect(ctx.Source())
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	_, err := ctx.Game.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      ctx.Controller(),
		Source:       ctx.Source(),
		Question:     "Flitting Guerrilla — exile this card? When you do, put target creature or battle card from your graveyard on top of your library.",
		AcceptLabel:  "Exile it",
		DeclineLabel: "Leave it",
		Asked:        game.NewAskedItem(ctx.Item),
	}, s82MomB12GuerrillaExileAnswer, nil)
	return err
})

// s82MomB12GuerrillaExileAnswer is the yes/no: accepting exiles the card, and
// the reflexive trigger follows only if it reached exile. An on-disk key.
var s82MomB12GuerrillaExileAnswer = game.RegisterConfirmThen("confirm/s82-mom-b12-guerrilla-exile", func(g *game.Game, r game.ConfirmAnswer) error {
	if !r.Accepted || r.Item == nil {
		return nil
	}
	ctx := NewContext(g, r.Item)
	z := g.FindCardZoneForEffect(r.Source)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{Target: r.Source, ThenTail: ctx.ItemTail(s82MomB12GuerrillaExiled)}.Apply(ctx)
})

// s82MomB12GuerrillaExiled is "when you do": only a card that really reached
// exile starts the reflexive trigger. An on-disk key.
var s82MomB12GuerrillaExiled = CardTail("tail/s82-mom-b12-guerrilla-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return WhenYouDo("Flitting Guerrilla — put target creature or battle card from your graveyard on top of your library",
		s82MomB12GuerrillaTopBody).Apply(ctx)
})
