package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fantastic Bounce — Sorcery {3}{U} (Marvel Super Heroes Commander,
// #4455):
//
//	"This spell costs {2} less to cast if it targets a tapped creature.
//	 Return target nonland permanent to its owner's hand.
//	 Draw a card."
//
// The reduction reads the announced target (CostsLessIfItTargets, like
// Price of Fame's), so the cast and the legal-move enumerator both price
// it with the target chosen. The draw waits behind the bounce
// (BounceToHand's ThenTail), so a replacement or an order prompt on the
// return is answered before the card is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71d09e00-e526-4899-beb3-0ee6d9b87f44",
		Name:         "Fantastic Bounce",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a tapped creature.",
				effects.And(effects.Creature(), msc7Tapped)),
		},
		OnResolve: effects.BounceTheTargetThenDraw,
	})
}

// msc7Tapped passes for a tapped permanent.
func msc7Tapped(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped }
