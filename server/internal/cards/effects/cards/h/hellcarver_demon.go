package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellcarver Demon — Creature — Demon {3}{B}{B}{B}, 6/6 (Rise of the Eldrazi,
// S82 OLD batch 483, #5581):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, sacrifice all
//	 other permanents you control and discard your hand. Exile the top six
//	 cards of your library. You may cast any number of spells from among cards
//	 exiled this way without paying their mana costs."
//
// The sentences run in order, each after the last has finished: the other
// permanents are sacrificed together (CR 701.21), then the hand is
// discarded, then six cards are exiled off the top, and the casts are made
// as part of this ability's resolution (CR 608.2g): a free cast of each
// exiled spell the controller picks, and a card left uncast stays in exile.
// Lands cannot be cast and stay in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "90514aa5-84b6-4be7-b6d0-b05529c97140",
		Name:            "Hellcarver Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Hellcarver Demon — sacrifice all other permanents you control and discard your hand; exile the top six cards; cast any number of them for free",
				hellcarverDemonBody),
		},
	})
}

// hellcarverDemonBody sacrifices every other permanent its controller
// controls, then carries on in the tail.
func hellcarverDemonBody(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.InstanceID != item.SourceCardID {
			ids = append(ids, c.InstanceID)
		}
	}
	return g.SacrificeAllKeyedForEffect(ctx.Source(), ids, ctx.Tail(hellcarverDemonSacrificed))
}

// hellcarverDemonSacrificed is the rest of the ability once the sacrifice
// has finished: discard the hand, then exile the top six. An on-disk key:
// never renamed, never reused.
var hellcarverDemonSacrificed = effects.CardTail("tail/hellcarver-demon-sacrificed", func(ctx *effects.Context, _ game.TailOutcome) error {
	if _, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller()); err != nil {
		return err
	}
	return ctx.Game.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 6, ctx.Tail(hellcarverDemonExiled))
})

// hellcarverDemonExiled offers the free casts over the cards that reached
// exile. An on-disk key: never renamed, never reused.
var hellcarverDemonExiled = effects.CardTail("tail/hellcarver-demon-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	effects.GrantFreeCasts(ctx.Game, ctx.Controller(), ctx.Source(), "Hellcarver Demon", game.TimingFlash, game.LapseStaysInExile, o.IDs)
	return nil
})
