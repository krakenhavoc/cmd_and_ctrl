package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cemetery Tampering — Enchantment {2}{B} (Duskmourn Commander, #5368):
//
//	"Hideaway 5 (When this enchantment enters, look at the top five cards
//	 of your library, exile one face down, then put the rest on the bottom
//	 in a random order.)
//	 At the beginning of your upkeep, you may mill three cards. Then if
//	 there are twenty or more cards in your graveyard, you may play the
//	 exiled card without paying its mana cost."
//
// Hideaway 5 is the shared Hideaway. The upkeep trigger captures the
// Tampering OBJECT in Build (CR 607.2a), so the card it plays is the one
// THIS incarnation hid. "You may mill" is a keyed yes/no; the mill is the
// "moved" form (CR 701.17c) and the count and the free play are its tail,
// so milled cards are in the graveyard before they are counted. "Then if"
// is not "if you do": a controller who declines to mill still plays the
// exiled card when their graveyard already holds twenty. The free play is
// PlayHiddenCard's, a land played during the resolution and a spell a {0}
// grant (ADR 0091).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9babcd0a-531f-417d-905e-b916a87112bb",
		Name:         "Cemetery Tampering",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Hideaway("Cemetery Tampering", 5),
			{
				Watches:   []game.EventKind{game.EventBeginUpkeep},
				AppliesTo: ByYou,
				Key:       cemeteryTamperingLabel,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, cemeteryTamperingLabel)
					item.Params.Object = game.ObjectRef(game.ObjectRefOf(*source))
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Cemetery Tampering — mill three cards?",
						Then:     cemeteryTamperingMayMill,
						Objects:  []game.ObjectRef{item.Params.Object},
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

const cemeteryTamperingLabel = "Cemetery Tampering — you may mill three cards; then if there are twenty or more cards in your graveyard, you may play the exiled card"

// cemeteryTamperingMayMill is the answer to "you may mill three cards":
// mill, then the graveyard count and the free play; or, declined, the count
// and the free play at once. An on-disk key: never renamed, never reused.
var cemeteryTamperingMayMill = MayAnswer("may/cemetery-tampering-mill", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Objects) == 0 {
		return nil
	}
	ref := game.PermissionCardRef(r.Objects[0])
	then := game.KeyedTail{Ref: cemeteryTamperingOffer, Player: ctx.Controller(), Source: ctx.Source(), Carry: []uuid.UUID{ref.ID}, N: ref.Epoch}
	if !r.Accepted {
		return then.Run(ctx.Game, nil)
	}
	return ctx.Game.MillToZoneKeyedMovedForEffect(ctx.Controller(), 3, game.ZoneGraveyard, then)
})

// cemeteryTamperingOffer is "then if there are twenty or more cards in your
// graveyard, you may play the exiled card without paying its mana cost". An
// on-disk key: never renamed, never reused.
var cemeteryTamperingOffer = CardTail("cemetery-tampering/offer-the-exiled-card", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	g, controller := ctx.Game, o.Tail.Player
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Graveyard == nil || p.Graveyard.Size() < 20 {
		return nil
	}
	ref := game.PermissionCardRef{ID: o.Tail.Carry[0], Epoch: o.Tail.N}
	grantHiddenPlay(g, controller, ref, "Cemetery Tampering — play the exiled card without paying its mana cost")
	return nil
})
