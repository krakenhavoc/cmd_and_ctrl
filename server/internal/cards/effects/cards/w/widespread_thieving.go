package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Widespread Thieving — Enchantment {2}{R} (Streets of New Capenna):
//
//	"Hideaway 5 (When this enchantment enters, look at the top five cards
//	 of your library, exile one face down, then put the rest on the bottom
//	 in a random order.)
//	 Whenever you cast a multicolored spell, create a Treasure token. Then
//	 you may pay {W}{U}{B}{R}{G}. If you do, you may play the exiled card
//	 without paying its mana cost."
//
// The cast trigger is Mage Tower Referee's (colours read as the spell sits
// on the stack). The Treasure lands first, so it can help pay; the payment
// is Balthier and Fran's MayPay, and the hidden card is the one THIS
// Widespread Thieving hid, carried as the object captured as the trigger was
// built (CR 607.2a). The free play is grantHiddenPlay (ADR 0091).
//
// No simplification.
const widespreadThievingLabel = "Widespread Thieving — create a Treasure, then you may pay {W}{U}{B}{R}{G} to play the exiled card"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c2e4b19-ac88-426a-9f1b-ac305cd512e8",
		Name:         "Widespread Thieving",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Hideaway("Widespread Thieving", 5),
			{
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventCast && ev.Actor == source.Controller && effects.B25CastIsMulticolored(ev, g)
				},
				Key: widespreadThievingLabel,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, widespreadThievingLabel)
					item.Params.Object = game.ObjectRef(game.ObjectRefOf(*source))
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CreateToken{Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.MayPay{
						Chooser:      item.Controller,
						Cost:         "{W}{U}{B}{R}{G}",
						Question:     "Widespread Thieving — pay {W}{U}{B}{R}{G}? If you do, you may play the exiled card without paying its mana cost.",
						Then:         widespreadThievingPaid,
						CarryObjects: []game.ObjectRef{item.Params.Object},
					}.Apply(ctx)
				},
			},
		},
	})
}

// widespreadThievingPaid is the paid branch: the hidden card is offered
// for a free play. An on-disk key.
var widespreadThievingPaid = effects.PayUnlessBranches("pay-unless/widespread-thieving-play",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) == 0 {
			return nil
		}
		effects.GrantHiddenPlay(ctx.Game, r.Chooser, game.PermissionCardRef(r.Objects[0]), "Widespread Thieving — play the exiled card without paying its mana cost")
		return nil
	}, nil)
