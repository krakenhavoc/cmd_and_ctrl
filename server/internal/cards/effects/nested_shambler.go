package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nested Shambler — Creature — Zombie {B}, 1/1 (EDHREC rank 4125):
//
//	"When this creature dies, create X tapped 1/1 green Squirrel
//	 creature tokens, where X is this creature's power."
//
// A one-mana body that refuses to stop being a body. On its own it
// trades and leaves a Squirrel behind; in the deck that actually
// plays it — a Chatterfang or a Slimefoot list with counters and
// anthems — the Shambler dies as a 5/5 and leaves five. It is a
// sacrifice-outlet engine that pays in wide, not tall.
//
// "THIS CREATURE'S POWER" IS THE LAST-KNOWN POWER, not the printed 1
// and not the power of the 1/1 card now lying in the graveyard.
// CR 603.10 / CR 608.2 read a dies-trigger's own characteristics off
// the game state just before it left the battlefield, which means the
// +1/+1 counters it was wearing and the anthem that was pumping it
// both count. b13LastKnownPower is that read: the layer-applied
// power from the LKI snapshot, plus the counters the event log says
// it held. A Shambler with three +1/+1 counters under a Glorious
// Anthem makes five Squirrels.
//
// The tokens ENTER TAPPED, as printed — they cannot block the turn
// the Shambler chumped, which is the drawback that makes the card
// cost one mana. They are green Squirrels and not black Zombies, so
// they feed a Chatterfang and not a Zombie lord; that mismatch is on
// the card, not on this file.
//
// X of zero or less makes nothing, which is what a Shambler killed by
// a -X/-X effect deserves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bfa882ee-de18-4ef8-8957-3e04ed6e3c1e",
		Name:         "Nested Shambler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Key: "Nested Shambler — a tapped Squirrel for each point of its power",
			Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				power := b13LastKnownPower(g, source.InstanceID, lki)
				return game.NewTriggeredItem(source, "Nested Shambler — a tapped Squirrel for each point of its power",
					func(g *game.Game, item *game.StackItem) error {
						if power <= 0 {
							return nil
						}
						return CreateTokenAdvanced{
							Controller: item.Controller,
							Spec:       Token(TokenCard("1/1 green Squirrel")).EntersTapped(),
							N:          power,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
