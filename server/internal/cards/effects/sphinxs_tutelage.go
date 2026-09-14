package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx's Tutelage — Enchantment {2}{U} (EDHREC rank 3654):
//
//	"Whenever you draw a card, target opponent mills two cards. If
//	 two nonland cards that share a color were milled this way,
//	 repeat this process.
//	 {5}{U}: Draw a card, then discard a card."
//
// The mono-blue mill engine. The draw trigger is Niv-Mizzet's shape
// — EventDrawCard fires once per card with the drawer in Actor, so a
// draw-three is three triggers, each with its own "target opponent"
// pick as it goes on the stack — and the body is the printed loop:
// the chosen opponent mills two; if both cards are nonland and share
// a colour (printed colours — Scryfall's stamped list, or the mana
// cost for a fixture), they mill two more, and so on until a pair
// fails the test or the library runs out. Each pass is bounded to
// the library, since a mill never loses a player the game (CR
// 704.5b). The loot is a CR 602 activation with a mana cost, and the
// drawn card fires the trigger again, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "158db438-52ba-4c30-b7bf-d1c4f76da917",
		Name:         "Sphinx's Tutelage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{5}{U}: Draw a card, then discard a card.",
			Cost:   ManaCost("{5}{U}"),
			Effect: b35LootOne,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, b35TutelageLabel, b35TutelageMill)
			},
		}},
	})
}
