package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malcolm, Keen-Eyed Navigator — 2/2 Legendary Creature — Siren
// Pirate for {2}{U}:
//
//	"Flying
//	 Whenever one or more Pirates you control deal damage to your
//	 opponents, you create a Treasure token for each opponent dealt
//	 damage.
//	 Partner"
//
// Breeches, Brazen Plunderer's trigger paying out in mana instead of
// cards — the ramp half of the same attack. Both partners in the
// same deck is the intended line.
//
// Batching (CR 603.1): one Treasure per damage event rather than one
// per opponent per combat. Three Pirates into three players gives
// three Treasures either way; three Pirates into the SAME player
// gives three here and one in paper.
//
// Partner is a deck-construction rule, not a game action, and is
// not modelled.
func init() {
	Register(Spec{
		OracleID:        "a66f8b44-0163-4456-b152-4acefab896a4",
		Name:            "Malcolm, Keen-Eyed Navigator",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Two Pirates hitting the same opponent make two Treasures instead of one; Partner isn't supported, so Malcolm can't be your commander."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if damagedOpponent(ev, source.Controller, g) == uuid.Nil {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && isPirate(dealer)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Malcolm — create a Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
