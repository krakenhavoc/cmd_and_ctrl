package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Town Greeter — Creature — Human Citizen {1}{G}, 1/1 (Final Fantasy,
// #4462):
//
//	"When this creature enters, mill four cards. You may put a land card
//	 from among them into your hand. If you put a Town card into your
//	 hand this way, you gain 2 life."
//
// Ainok Wayfarer's mill-then-take shape (mill_then_take.go): only the
// cards this mill put into the graveyard are offered, filtered to lands,
// and the pick is optional. The life is gained only when the card taken
// really reached the hand and is a Town, read as it arrives; declining,
// or milling no land, gains nothing. Neither half depends on the Greeter
// still being on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9dc61ef6-da22-42a8-b348-dfd5722deadf",
		Name:         "Town Greeter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Town Greeter — mill four cards; you may put a land card from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 4, ThenTail: ctx.ItemTail(townGreeterMilled)}.Apply(ctx)
				}),
		},
	})
}

// townGreeterMilled offers a land from among the milled cards. An
// on-disk key.
var townGreeterMilled = effects.CardTail("town-greeter/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	player := ctx.Controller()
	var lands []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !effects.Land()(ctx.Game, player, c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		lands = append(lands, id)
	}
	if len(lands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Town Greeter — you may put a land card from among them into your hand (a Town gains you 2 life)",
		Cards:    lands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  townGreeterPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// townGreeterPick puts the chosen land into the hand, and gains 2 life
// if it is a Town that got there. An on-disk key.
var townGreeterPick = effects.AskedPick("choose-cards/town-greeter-land", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
		return err
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.HasSubtype("Town") {
		return effects.GainLife{Amount: 2}.Apply(ctx)
	}
	return nil
})
