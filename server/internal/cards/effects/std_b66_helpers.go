package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b66_helpers.go — shared bodies for Lorwyn Eclipsed leftovers, batch 66
// (#4458).

// stdB66EndBlazeBody is End-Blaze Epiphany's delayed trigger: "exile a
// number of cards from the top of your library equal to its power, then
// choose a card exiled this way. Until the end of your next turn, you may
// play that card." The power is the creature's as it last existed on the
// battlefield (CR 608.2h), read from the object the trigger was created
// about. A creature with no power, or an empty library, exiles nothing.
var stdB66EndBlazeBody = game.SimpleDelayedBody("end-blaze-epiphany/exile-and-choose", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	info, ok := g.PermanentForEffect(obj)
	if !ok || info.Power <= 0 {
		return nil
	}
	ids := g.LookAtTopOfLibraryForEffect(item.Controller, info.Power)
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(stdB66EndBlazeChoose))
})

// stdB66EndBlazeChoose is "choose a card exiled this way" over the cards that
// reached exile. An on-disk key.
var stdB66EndBlazeChoose = CardTail("then/end-blaze-epiphany-choose", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "End-Blaze Epiphany — choose a card exiled this way (you may play it until the end of your next turn)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  stdB66EndBlazeGrant,
	})
	return nil
})

// stdB66EndBlazeGrant gives the chosen card its window. An on-disk key.
var stdB66EndBlazeGrant = game.RegisterChooseCardsThen("choose-cards/end-blaze-epiphany-play", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfYourNextTurnDuration(r.Chooser),
			Label:    "End-Blaze Epiphany — play it until the end of your next turn",
		})
	}
	return nil
})

// stdB66MeekAttackSacrifice is Meek Attack's delayed trigger: sacrifice the
// creature it put onto the battlefield, but only if it is still that object
// and still under the controller's control (the ruling, CR 603.7c): a
// creature that left and came back is a new object and is not sacrificed.
var stdB66MeekAttackSacrifice = game.DelayedBody("meek-attack/sacrifice", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || !onBattlefield(g, p.Object.ID) || c.ObjectEpoch != p.Object.Epoch || c.Controller != item.Controller {
		return nil
	}
	return SacrificePermanent{Target: p.Object.ID}.Apply(NewContext(g, item))
})

// stdB66DamageSourceHasType is "a source of the chosen type": the damage
// source as it is (or last was), a permanent, spell or other object with that
// creature type, a changeling being every type (CR 702.73a).
func stdB66DamageSourceHasType(ch game.Characteristic, tribe string) bool {
	if tribe == "" {
		return false
	}
	return slices.Contains(ch.Subtypes, tribe) || (ch.AllCreatureTypes && game.IsCreatureType(tribe))
}
