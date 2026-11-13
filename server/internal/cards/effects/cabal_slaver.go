package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cabal Slaver — Creature — Human Cleric {2}{B}, 2/1 (Onslaught):
//
//	"Whenever a Goblin deals combat damage to a player, that player
//	 discards a card."
//
// Any Goblin's combat damage, whoever controls it and whoever it hits
// (the Slaver's controller included), one trigger per Goblin that
// connects. "Goblin" is read off the creature as the damage is dealt,
// through the Slaver's own text (CR 612.2), so a changeling counts. "That
// player" is the damage event's target and chooses the card (CR 701.9b),
// Abyssal Specter's discardOne; an empty hand discards nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22a189a9-e755-4cbb-b419-5ba2e5acbe1b",
		Name:         "Cabal Slaver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, cabalSlaverAGoblinDealtCombatDamageToAPlayer,
				"Cabal Slaver — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return discardOne(g, item, NewContext(g, item).Trigger().Event.Target)
				}),
		},
	})
}

// cabalSlaverAGoblinDealtCombatDamageToAPlayer is "whenever a Goblin
// deals combat damage to a player": combat damage to a seated player from
// a creature, any player's, that is a Goblin as the damage is dealt.
func cabalSlaverAGoblinDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && OfCreatureType("Goblin")(g, source.Controller, game.ThroughTextOf(source, c))
}
