package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sin Prodder — Creature — Devil {2}{R}, 3/2 (Jumpstart, #5510):
//
//	"Menace
//	 At the beginning of your upkeep, reveal the top card of your library.
//	 Any opponent may have you put that card into your graveyard. If a
//	 player does, this creature deals damage to that player equal to that
//	 card's mana value. Otherwise, put that card into your hand."
//
// The top card is revealed to the table (CR 701.20). Distant Memories'
// offer (ADR 0150) then goes to each opponent in APNAP order from the
// active player, the controller skipped, until one says yes at no cost. A
// yes puts the revealed card into the graveyard and the Prodder deals
// damage to that player equal to its mana value (an X is 0, CR 107.3m); if
// nobody says yes the card goes to the controller's hand. With an empty
// library nothing is revealed and nothing happens. A card that has left
// the library by then is left where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "745efa76-7063-4971-9d1b-b2b4cb336a57",
		Name:            "Sin Prodder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Sin Prodder — reveal the top card of your library; any opponent may have you put it into your graveyard", func(g *game.Game, item *game.StackItem) error {
				revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1, "Sin Prodder — reveal the top card of your library")
				if len(revealed) == 0 {
					return nil
				}
				return effects.AnyPlayerPays{
					Pay:          effects.NoPayment(),
					Opponents:    true,
					Question:     "Sin Prodder — put the revealed card into its controller's graveyard? If you do, Sin Prodder deals damage to you equal to its mana value; otherwise its controller puts it into their hand.",
					AcceptLabel:  "Put it into their graveyard",
					DeclineLabel: "Let them have it",
					Then:         sinProdderOutcome,
					Carry:        revealed,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// sinProdderOutcome is both answers, keyed (#4264): the revealed card rides
// the offer. A yes puts it into the graveyard and the Prodder hits the
// player who said yes; nobody saying yes puts it into its owner's hand.
var sinProdderOutcome = effects.AnyPlayerConsequences("any-player/sin-prodder-outcome",
	func(ctx *effects.Context, r game.AnyPlayerAnswer) error {
		for _, id := range r.Carry {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !inLibrary(ctx, id) {
				continue
			}
			mv := c.ManaValue()
			if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
				return err
			}
			// "If a player does": the damage follows the yes, not the
			// card's arrival, so a replaced graveyard move still pays out.
			return effects.DealDamage{Source: ctx.Source(), Target: r.Payer, Amount: mv}.Apply(ctx)
		}
		return nil
	},
	func(ctx *effects.Context, r game.AnyPlayerAnswer) error {
		for _, id := range r.Carry {
			if inLibrary(ctx, id) {
				if err := ctx.Game.BounceToHandForEffect(id); err != nil {
					return err
				}
			}
		}
		return nil
	},
	nil)

// inLibrary reports whether the card is still in a library.
func inLibrary(ctx *effects.Context, id uuid.UUID) bool {
	z := ctx.Game.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneLibrary
}
