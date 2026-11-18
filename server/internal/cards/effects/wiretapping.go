package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wiretapping — Enchantment {4}{U} (Streets of New Capenna):
//
//	"Hideaway 5 (When this enchantment enters, look at the top five cards
//	 of your library, exile one face down, then put the rest on the bottom
//	 in a random order.)
//	 Whenever you draw your first card during each of your draw steps, draw
//	 a card. Then if you have nine or more cards in hand, you may play the
//	 exiled card without paying its mana cost."
//
// The trigger is a draw event of yours that is the first completed draw of
// your own draw step (FirstCompletedDrawInOwnDrawStepForEffect, Orcish
// Bowmasters' reader); the extra draw is itself a draw, but the step's tally
// is two by then, so it does not trigger again. The hidden card is the one
// THIS Wiretapping hid, captured as the object as the trigger is built
// (CR 607.2a). The nine-card check is the draw's continuation, so it reads
// the hand after the card has really landed even when the draw pauses on a
// prompt. The free play is grantHiddenPlay (ADR 0091).
//
// No simplification.
const wiretappingLabel = "Wiretapping — draw a card, then if you have nine or more cards in hand, you may play the exiled card"

func init() {
	Register(Spec{
		OracleID:     "befc5986-f06c-40a4-98ee-ffcace4ceaa8",
		Name:         "Wiretapping",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Hideaway("Wiretapping", 5),
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && g.FirstCompletedDrawInOwnDrawStepForEffect(ev.Actor)
				},
				Key: wiretappingLabel,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, wiretappingLabel)
					item.Params.Object = game.ObjectRef(game.ObjectRefOf(*source))
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ref := game.PermissionCardRef(item.Params.Object)
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref:    wiretappingAfterDraw,
						Player: item.Controller,
						Source: ref.ID,
						N:      ref.Epoch,
					})
				},
			},
		},
	})
}

// wiretappingAfterDraw is "then if you have nine or more cards in hand, you
// may play the exiled card": Source is the hideaway object's ID and N its
// epoch. An on-disk key: never renamed, never reused.
var wiretappingAfterDraw = game.RegisterDrawThen("wiretapping/after-draw", func(g *game.Game, d game.DrawThen) error {
	if b14HandSize(g, d.Player) < 9 {
		return nil
	}
	grantHiddenPlay(g, d.Player, game.PermissionCardRef{ID: d.Source, Epoch: d.N}, "Wiretapping — play the exiled card without paying its mana cost")
	return nil
})
