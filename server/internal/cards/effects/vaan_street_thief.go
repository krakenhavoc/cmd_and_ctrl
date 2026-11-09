package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vaan, Street Thief — Legendary Creature — Human Scout {2}{R}, 2/2 (Final
// Fantasy, #4462):
//
//	"Whenever one or more Scouts, Pirates, and/or Rogues you control deal
//	 combat damage to a player, exile the top card of that player's
//	 library. You may cast it. If you don't, create a Treasure token.
//	 Whenever you cast a spell you don't own, put a +1/+1 counter on each
//	 Scout, Pirate, and Rogue you control."
//
// The first ability is a batch clause (CR 603.2c): one trigger per player
// those creatures connect with in a combat damage step. The exile lands, then
// the cast is offered inside the resolution (CR 608.2g) at any timing, paying
// the card's mana cost, with the permission lapsing when the offer is
// answered; whatever is not cast (a decline, a land, a card that can't be
// paid for, an empty library) makes the Treasure. The second counts the
// spell's owner, not who cast it, so a card cast from an opponent's library or
// exile counts and your own never does; each creature of the three types you
// control gets its counter as the ability resolves, Vaan himself a Scout
// among them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24fd1f47-bb15-45c6-a7b1-ba8eb3bb8b1d",
		Name:         "Vaan, Street Thief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				Or(OfCreatureType("Scout"), OfCreatureType("Pirate"), OfCreatureType("Rogue")),
				"Vaan, Street Thief — exile the top card of that player's library; you may cast it; if you don't, create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					victim := item.Trigger.Event.Target
					if g.PlayerByIDForEffect(victim) != nil {
						if top := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), victim, 1); len(top) > 0 {
							return g.ExileCardKeyedForEffect(top[0], ctx.ItemTail(vaanExiled))
						}
					}
					// Nothing to exile, so nothing was cast.
					return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
				}),
			WheneverYouCast(Not(YouOwn()), "Vaan, Street Thief — put a +1/+1 counter on each Scout, Pirate, and Rogue you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() &&
							(c.HasSubtype("Scout") || c.HasSubtype("Pirate") || c.HasSubtype("Rogue")) {
							ids = append(ids, c.InstanceID)
						}
					}
					return avB24PutCountersOnEachThen(ctx, ids, game.CounterPlusOne, 1, game.KeyedTail{})
				}),
		},
	})
}

// vaanExiled is the exile landed: the cast is offered over the card, and the
// Treasure follows unless it was cast. An on-disk key.
var vaanExiled = CardTail("vaan-street-thief/exiled", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	if len(o.IDs) == 0 {
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	}
	perm := game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  "Vaan, Street Thief",
		Label:       "Vaan, Street Thief — cast it",
	}
	landed := false
	if g.GrantCastPermissionOverCardForEffect(o.IDs[0], perm) {
		landed = g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
			Chooser:      ctx.Controller(),
			Source:       ctx.Source(),
			Reason:       "Vaan, Street Thief",
			Cards:        o.IDs,
			Keyword:      game.MayCastKeywordPaidCast,
			AcceptLabel:  "Cast it",
			DeclineLabel: declineLeaveInExile,
			Then:         vaanAfterCast,
			Carry:        o.IDs,
		})
	}
	if landed {
		return nil
	}
	return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
})

// vaanAfterCast is the offer answered: no card cast means a Treasure. An
// on-disk key.
var vaanAfterCast = game.RegisterChooseCardsThen("resolution-cast/vaan-street-thief-after-cast",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) > 0 {
			return nil
		}
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	})
