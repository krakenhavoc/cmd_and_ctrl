package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Mourning — Creature — Wall {1}{W}, 0/4 (Midnight Hunt Commander,
// S82 MID batch 13):
//
//	"Defender
//	 When this creature enters, exile a card from the top of your library
//	 face down for each opponent you have.
//	 Coven — At the beginning of your end step, if you control three or more
//	 creatures with different powers, put a card exiled with this creature
//	 into its owner's hand."
//
// The enters trigger exiles one card per opponent still in the game, face down
// (CR 406.3a), and links each to this object (CR 607.2a) once it has landed:
// a Wall that leaves and returns is a new object with no claim on them (CR
// 400.7). The end-step trigger is an intervening "if" (CR 603.4) on three
// distinct powers through the layers (trekB48Coven), checked as the step
// begins and again as it resolves. It puts one card exiled with the Wall into
// its owner's hand: with one linked card that card, with several the
// controller picks (the cards are face down, so the pick is blind). Nothing
// linked, nothing happens.
//
// No simplification.
func init() {
	coven := effects.CondBoard{
		Name: "you control three or more creatures with different powers",
		Test: func(ctx *effects.Context) bool { return effects.TrekB48Coven(ctx.Game, ctx.Controller(), ctx.Source()) },
	}
	effects.Register(effects.Spec{
		OracleID:        "585c9671-7451-4506-8b21-0d64106c4cd7",
		Name:            "Wall of Mourning",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wall of Mourning — exile a card from the top of your library face down for each opponent you have",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := len(ctx.Opponents())
					if n == 0 {
						return nil
					}
					return g.ExileTopFaceDownKeyedForEffect(item.Controller, n, ctx.ItemTail(wallOfMourningLink))
				}),
			effects.InterveningIf(coven, effects.On(game.EventBeginEndStep, effects.ByYou,
				"Wall of Mourning — Coven: put a card exiled with this creature into its owner's hand",
				wallOfMourningReturn)),
		},
	})
}

// wallOfMourningLink links each card the exile took to the Wall (the
// asking item's source object). An on-disk key.
var wallOfMourningLink = effects.CardTail("then/wall-of-mourning-link", func(ctx *effects.Context, o game.TailOutcome) error {
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, id := range o.IDs {
		ctx.Game.RelinkExiledWithForEffect(id, game.PermissionCardRef(ref))
	}
	return nil
})

// wallOfMourningReturn is the end-step ability: the cards exiled with this
// object, one into its owner's hand.
func wallOfMourningReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return g.BounceToHandForEffect(ids[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Wall of Mourning — put a card exiled with this creature into its owner's hand",
		Cards:      ids,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneExile,
		ThenKey:    wallOfMourningPick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// wallOfMourningPick puts the chosen card into its owner's hand. An on-disk
// key.
var wallOfMourningPick = effects.AskedPick("choose-cards/wall-of-mourning-pick", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		return ctx.Game.BounceToHandForEffect(id)
	}
	return nil
})
