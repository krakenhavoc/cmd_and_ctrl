package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Furious Forebear — Creature — Spirit Warrior {1}{W}, 3/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Whenever a creature you control dies while this card is in your
//	 graveyard, you may pay {1}{W}. If you do, return this card from your
//	 graveyard to your hand."
//
// A trigger that works from the graveyard (CR 113.6, InGraveyard), on a
// creature you control dying. The card must already have been in the
// graveyard: Furious Forebear dying at the same time as creatures you
// control, or being one of them, does not trigger it (the rulings), which
// is read off the event batch: its own arrival in a graveyard shares the
// batch of the death that would trigger it. "You may pay {1}{W}. If you
// do" is a MayPay asked as the trigger resolves. The return needs the card
// to still be in the graveyard as the same object (CR 400.7); one that
// left before the ability resolved is not returned even when the cost was
// paid.
//
// No simplification.
func init() {
	t := On(game.EventLTB, furiousForebearTriggers, "Furious Forebear — you may pay {1}{W}; if you do, return this card from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}{W}",
				Question: "Furious Forebear — pay {1}{W} to return it from your graveyard to your hand?",
				Then:     furiousForebearPaid,
			}.Apply(NewContext(g, item))
		})
	Register(Spec{
		OracleID:     "4fb1d33f-a87d-46cf-a73e-882c3dbcb26f",
		Name:         "Furious Forebear",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{InGraveyard(t)},
	})
}

// furiousForebearTriggers is "a creature you control dies" while the card
// was already in the graveyard.
func furiousForebearTriggers(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	return ACreatureYouControlDied(ev, source, ch, g) && furiousForebearWasInGraveyard(ev, source, ch, g)
}

// furiousForebearWasInGraveyard is "while this card is in your graveyard":
// the card was already there before the death. It is false for the card's
// own death and for any death in the batch that put the card there.
func furiousForebearWasInGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		prior := g.Events[i]
		if prior.Batch < ev.Batch {
			break
		}
		if prior.Batch == ev.Batch && prior.Kind == game.EventLTB && prior.CardID == source.InstanceID &&
			prior.NewZone == game.ZoneGraveyard {
			return false
		}
	}
	return true
}

// furiousForebearPaid returns the card to its owner's hand if it is still
// in the graveyard as the object the ability came from. An on-disk key.
var furiousForebearPaid = PayUnlessBranches("pay-unless/furious-forebear-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		self := ctx.Source()
		z := ctx.Game.FindCardZoneForEffect(self)
		if self == uuid.Nil || z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		card, ok := ctx.Game.LookupCardForEffect(self)
		if !ok {
			return nil
		}
		if ref := ctx.Item.SourceObject; ref.ID == self && ref.Epoch != card.ObjectEpoch {
			return nil
		}
		return ReturnFromGraveyard{Target: self, Dest: game.ZoneHand}.Apply(ctx)
	}, nil)
