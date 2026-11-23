package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dollmaker's Shop // Porcelain Gallery — Enchantment — Room (ADR 0103):
//
//	Dollmaker's Shop {1}{W}: "Whenever one or more non-Toy creatures you
//	control attack a player, create a 1/1 white Toy artifact creature
//	token."
//	Porcelain Gallery {4}{W}{W}: "Creatures you control have base power
//	and toughness each equal to the number of creatures you control."
//
// The shop is the "one or more … attack A PLAYER" batch guard
// (OncePerBatchPerPlayer, CR 603.2c) narrowed to non-Toy attackers: a
// Toy declared alongside does not make a second Toy on its own. The
// gallery is a layer 7b set whose value is counted on every recompute;
// counters and anthems land on top of it in 7c and 7d.
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "e633412b-e36b-4ec5-b0d0-7cb24c7503f4",
		Name:         "Dollmaker's Shop // Porcelain Gallery",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.On(game.EventAttack, nonToyCreatureYouControlAttackedAPlayer,
				"Dollmaker's Shop — create a 1/1 white Toy artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Toy artifact"), N: 1}))),
		}},
		Right: effects.Door{Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: effects.B16CreaturesYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B39CreaturesControlledBy(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}}},
	}))
}

// nonToyCreatureYouControlAttackedAPlayer is Dollmaker's Shop's
// condition: an attack declared by you, at a player, by a creature that
// is not a Toy.
func nonToyCreatureYouControlAttackedAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.B16YouAttackedAPlayer(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	// "Toy" is a word of the source's text (CR 612.2, #3984).
	return ok && c.IsCreature() && !c.HasSubtype(source.TextSubtype("Toy"))
}
