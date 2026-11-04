package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow of the Goblin — Enchantment {1}{R} (Marvel's Spider-Man):
//
//	"Unreliable Visions — At the beginning of your first main phase,
//	 discard a card. If you do, draw a card.
//	 Undying Vengeance — Whenever you play a land or cast a spell from
//	 anywhere other than your hand, this enchantment deals 1 damage to each
//	 opponent."
//
// Unreliable Visions is mandatory: the discard is asked as the trigger
// resolves (an empty hand discards nothing), and the draw follows only from a
// card really discarded (CR 608.2c), the registered "draw one if any" rule
// the may-discard cards share. Undying Vengeance is one ability with two
// trigger conditions watching one declaration (The Endstone's shape): a cast
// whose origin zone is not a hand (Vega, the Watcher's read), or a land play
// from a zone that is not a hand (the settled entry's Played marker, CR
// 305.4). A copy of a spell is not cast, and a land put onto the battlefield
// without being played, or played from the hand, triggers nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f343b6f9-41c3-419b-a57c-a7b08603ef5c",
		Name:         "Shadow of the Goblin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Shadow of the Goblin — discard a card; if you do, draw a card",
				discardACardIfYouDoDraw),
			OnAny([]game.EventKind{game.EventCast, game.EventZoneMove},
				shadowOfTheGoblinPlayedFromOutsideHand,
				"Shadow of the Goblin — 1 damage to each opponent", Do(DamageEachOpponent{N: 1})),
		},
	})
}

// shadowOfTheGoblinPlayedFromOutsideHand is "you play a land or cast a spell
// from anywhere other than your hand".
func shadowOfTheGoblinPlayedFromOutsideHand(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller || ev.OldZone == "" || ev.OldZone == game.ZoneHand {
		return false
	}
	return ev.Kind == game.EventCast || b20LandPlayed(ev, g)
}
