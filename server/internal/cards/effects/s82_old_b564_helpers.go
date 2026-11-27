package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b564_helpers.go — Recover (CR 702.59a), Coldsnap's keyword:
//
//	"Recover [cost]" means "When a creature is put into your graveyard from
//	the battlefield, you may pay [cost]. If you do, return this card from your
//	graveyard to your hand. Otherwise, exile this card."
//
// A triggered ability that functions only while its card is in a graveyard.

// S82OldB564CreaturePutIntoYourGraveyard is Recover's trigger condition: a
// creature, as it last existed on the battlefield (CR 603.10a), went from the
// battlefield to the graveyard of the player who owns the card with recover.
// "You" for a card in a graveyard is its owner (CR 108.4). The card with
// recover is never the creature that did it, and it does not see a creature
// that went to the graveyard in the same simultaneous exit as it did: it was
// not in the graveyard before the event (CR 603.10a looks back for
// leaves-the-battlefield abilities only, and recover is not one).
func S82OldB564CreaturePutIntoYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID || g.InSameSimultaneousExitForEffect(ev.CardID, source.InstanceID) {
		return false
	}
	dead, ok := DiedCreature(ev, g)
	return ok && dead.Owner == source.Controller
}

// s82OldB564RecoverBranches is the keyed answer to Recover's "you may pay":
// paid, the card returns from the graveyard to its owner's hand; otherwise it
// is exiled. Either does nothing if the card has already left the graveyard
// (CR 400.7). An on-disk key: never renamed, never reused.
var s82OldB564RecoverBranches = PayUnlessBranches("pay-unless/s82-old-b564-recover",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReturnThisFromYourGraveyardToHand(ctx.Game, ctx.Item)
	},
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ExileFromGraveyardIfStillThere(ctx.Game, ctx.Item, ctx.Item.SourceCardID)
	})

// S82OldB564Recover is "Recover {cost}" for a mana cost: the trigger of a card
// in its owner's graveyard.
func S82OldB564Recover(card, cost string) game.TriggeredAbility {
	label := card + " — recover " + cost
	return InGraveyard(On(game.EventLTB, S82OldB564CreaturePutIntoYourGraveyard, label,
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{
				Chooser:  item.Controller,
				Cost:     cost,
				Question: card + " — pay " + cost + " to return it from your graveyard to your hand? (Otherwise it is exiled.)",
				Then:     s82OldB564RecoverBranches,
			}.Apply(NewContext(g, item))
		}))
}

// S82OldB564RecoverHalfLife is "Recover—Pay half your life, rounded up"
// (Garza's Assassin). The amount is worked out as the trigger resolves, from
// the owner's life then. A payment the owner cannot make at all (CR 118.3, a
// "can't pay life" effect) is not offered, and "otherwise" applies: the card
// is exiled.
func S82OldB564RecoverHalfLife(card string) game.TriggeredAbility {
	label := card + " — recover: pay half your life, rounded up"
	return InGraveyard(On(game.EventLTB, S82OldB564CreaturePutIntoYourGraveyard, label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			p := g.PlayerByIDForEffect(item.Controller)
			if p == nil {
				return nil
			}
			pay := LifePayment(max(p.Life+1, 0) / 2)
			if pay.Action.Count <= 0 || !g.PayActionCanPayForEffect(item.Controller, pay.Action) {
				return ExileFromGraveyardIfStillThere(g, item, item.SourceCardID)
			}
			return MayPay{
				Chooser:  item.Controller,
				Question: card + " — pay half your life, rounded up, to return it from your graveyard to your hand? (Otherwise it is exiled.)",
				Pay:      pay,
				Then:     s82OldB564RecoverBranches,
			}.Apply(ctx)
		}))
}
