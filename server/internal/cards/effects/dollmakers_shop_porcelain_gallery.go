package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Room(RoomSpec{
		OracleID:     "e633412b-e36b-4ec5-b0d0-7cb24c7503f4",
		Name:         "Dollmaker's Shop // Porcelain Gallery",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, nonToyCreatureYouControlAttackedAPlayer,
				"Dollmaker's Shop — create a 1/1 white Toy artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Toy artifact"), N: 1}))),
		}},
		Right: Door{Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: b16CreaturesYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b39CreaturesControlledBy(g, source.Controller)
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
	if !b16YouAttackedAPlayer(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && !c.HasSubtype("Toy")
}
