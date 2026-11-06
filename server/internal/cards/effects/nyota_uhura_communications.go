package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nyota Uhura, Communications — Legendary Creature — Human Officer
// {2}{U}, 2/4 (Star Trek, #4973):
//
//	"Hailing Frequencies — As Nyota Uhura enters, choose a number between
//	 0 and 10.
//	 Open a Channel — Whenever an opponent casts a spell with mana value
//	 equal to the chosen number, you draw a card."
//
// The number is a CR 614.12 "as enters" choice, asked of the entering
// permanent's controller and stored on it (game.Card.ChosenNumber, the
// slot Phyrexian Processor's life uses, through the same keyed branch).
// The trigger reads the spell's mana value on the stack, with any {X} at
// the value chosen for it (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d89f34e4-8628-4faf-89e6-c9c16023ae8b",
		Name:         "Nyota Uhura, Communications",
		Completeness: CompletenessFull,
		AsEnters:     nyotaUhuraChooseANumber,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, nyotaUhuraOpponentCastTheNumber,
				"Open a Channel — an opponent cast a spell with the chosen mana value: you draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}

// nyotaUhuraChooseANumber is Hailing Frequencies: a number from 0 to 10,
// stored on the permanent as it enters.
func nyotaUhuraChooseANumber(card *game.Card, ctx *Context) error {
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Resource: game.PayResourceNone,
		Min:      0,
		Max:      10,
		Goal:     3,
		Question: "Hailing Frequencies — choose a number between 0 and 10",
		Branch:   game.NumberBranch{Then: storeChosenNumber, Carry: []uuid.UUID{card.Controller}},
	})
}

// nyotaUhuraOpponentCastTheNumber is "Whenever an opponent casts a spell
// with mana value equal to the chosen number".
func nyotaUhuraOpponentCastTheNumber(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !AnOpponentCast(nil)(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	mv, ok := g.ManaValueForEffect(c)
	return ok && mv == source.ChosenNumber
}
