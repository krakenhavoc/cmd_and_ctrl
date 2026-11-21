package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gishath, Sun's Avatar — Legendary Creature — Dinosaur Avatar
// {5}{R}{G}{W}, 7/6 (The Lost Caverns of Ixalan, #4470):
//
//	"Vigilance, trample, haste
//	 Whenever Gishath deals combat damage to a player, reveal that many
//	 cards from the top of your library. Put any number of Dinosaur
//	 creature cards from among them onto the battlefield and the rest on
//	 the bottom of your library in a random order."
//
// "That many" is the damage the trigger saw (April, Reporter of the
// Weird's read of the event). The cards are revealed to every player,
// then Earth's Mightiest Heroes' "any number" put (Max 0, Optional:
// any number includes none) offers the Dinosaur creature cards among
// them; the picks enter as one simultaneous batch. Everything not put
// goes to the bottom in a random order (PutRestOnBottomInRandomOrderTail).
// "Dinosaur" is read through the trigger's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a311daab-40dc-4433-9ed8-feab7c558bbf",
		Name:            "Gishath, Sun's Avatar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Gishath, Sun's Avatar — reveal that many cards and put any number of Dinosaur creature cards onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					n := item.Trigger.Event.Amount
					if n <= 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), n, "Gishath, Sun's Avatar — revealed from the top of the library"),
						Match:    effects.And(effects.Creature(), effects.HasSubtype(ctx.TextSubtype("Dinosaur"))),
						Optional: true,
						Label:    "Gishath, Sun's Avatar — put any number of Dinosaur creature cards onto the battlefield",
						ThenTail: effects.PutRestOnBottomInRandomOrderTail,
					}.Apply(ctx)
				}),
		},
	})
}
