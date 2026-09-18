package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Open the Graves — Enchantment {3}{B}{B} (EDHREC rank 3856):
//
//	"Whenever a nontoken creature you control dies, create a 2/2
//	 black Zombie creature token."
//
// Midnight Reaper's trigger with a body on the end instead of a card
// — the same EventLTB read, the same nontoken guard, and the guard is
// the whole card: without it the Zombie it makes would itself die and
// make another, and a single sacrifice outlet would be an infinite
// loop. The printed word "nontoken" is the anti-combo clause, not
// flavour, which is why it is checked rather than assumed.
//
// Reads the dead creature POST-MOVE, the diedCreature posture: the
// controller compared against this enchantment's is the one the
// creature had as it left the battlefield (CR 603.10 last-known
// information), so a creature stolen by an opponent and killed there
// does not feed its original owner's Graves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28778958-a1f9-4fea-b551-c193d1257f18",
		Name:         "Open the Graves",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller == source.Controller && !IsToken(dead)
			}, "Open the Graves — create a 2/2 black Zombie", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("2/2 black Zombie"),
					N:          1,
				}.Apply(ctx)
			}),
		},
	})
}
