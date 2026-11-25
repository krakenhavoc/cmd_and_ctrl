package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foul Emissary — Creature — Human Horror {2}{G}, 1/1 (Eldritch Moon, S82
// OLD, #5541):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal a creature card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in any order.
//	 When you sacrifice this creature while casting a spell with emerge,
//	 create a 3/2 colorless Eldrazi Horror creature token."
//
// The enters trigger is Courageous Outrider's dig with "creature card". The
// second trigger watches the Emissary's own sacrifice as a cost while its
// controller has a spell on the stack that was cast with emerge: the
// sacrifice is paid with the spell already cast (CR 601.2h), so the trigger
// goes on the stack above it and resolves first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "034c2b40-6bcc-4bb5-be94-4602496ddcc6",
		Name:         "Foul Emissary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Foul Emissary — look at the top four cards of your library; you may reveal a creature card from among them and put it into your hand; put the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Match:    effects.Creature(),
						Max:      1,
						Optional: true,
						Label:    "Foul Emissary — you may reveal a creature card from among them and put it into your hand",
						Key:      effects.TakeRevealRestBottomAny,
					}.Apply(ctx)
				}),
			effects.On(game.EventSacrifice, effects.S82OldB277SacrificedWhileCastingEmerge,
				"Foul Emissary — create a 3/2 colorless Eldrazi Horror creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: 1})),
		},
	})
}
