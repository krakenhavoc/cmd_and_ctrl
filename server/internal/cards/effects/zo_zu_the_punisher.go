package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zo-Zu the Punisher — Legendary Creature — Goblin Warrior {1}{R}{R},
// 2/2 (EDHREC rank 4055):
//
//	"Whenever a land enters, Zo-Zu deals 2 damage to that land's
//	 controller."
//
// The symmetrical land tax, and the reason it is played in Commander
// at all: a four-player pod makes three land drops a turn cycle that
// are not yours, and each of them is two damage. Zo-Zu hurts his own
// controller too — the ability says "a land", full stop — which is
// the honest half of the card and is modelled as printed.
//
// EVERY LAND, NOT JUST AN OPPONENT'S, AND NOT JUST A LAND DROP. The
// trigger watches the ETB event, so a land put onto the battlefield
// by a fetch, a Cultivate, an Oracle of Mul Daya reveal or a token
// Dryad fires it exactly as a hand land drop does — CR 603.2, "enters
// the battlefield" is the event, not "is played". That is what makes
// Zo-Zu punishing rather than merely annoying against a ramp deck.
//
// "THAT LAND'S CONTROLLER" is read off the entering card, not off the
// active player: a land that enters under an opponent's control on
// your turn (a Hunting Wilds, a stolen fetch) damages THEM. The
// controller is read as the trigger is built, so a land that has
// since left is still attributed correctly.
//
// The damage is Zo-Zu's, which means it is red noncombat damage from
// a creature source: it is doubled by Fiery Emancipation, prevented
// by a fog on that player, and reflected by anything that watches
// damage from creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e412b68-9179-4094-960a-95692428855b",
		Name:         "Zo-Zu the Punisher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b39LandEnteredController(ev, g)
				return ok
			},
			Key: "Zo-Zu the Punisher — 2 damage to that land's controller",
			// A fill-in Build (ADR 0041 P9): the controller is read
			// as the trigger is built, so it is captured into
			// Params.Player rather than a closure — a land that has
			// since left, or changed controller, is still attributed
			// correctly.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Zo-Zu the Punisher — 2 damage to that land's controller")
				item.Params.Player, _ = b39LandEnteredController(ev, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDamage{
					Source: item.SourceCardID,
					Target: item.Params.Player,
					Amount: 2,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
