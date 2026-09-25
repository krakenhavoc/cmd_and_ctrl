package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raid Bombardment — Enchantment {2}{R} (EDHREC rank 3476):
//
//	"Whenever a creature you control with power 2 or less attacks,
//	 this enchantment deals 1 damage to the player or planeswalker
//	 that creature is attacking."
//
// The go-wide deck's reach. One trigger per small attacker — the
// engine emits one EventAttack per declared creature, exactly the
// printed once-per-creature shape — with the power read as the
// creature is declared (counters and anthems included), and the
// damage from the enchantment to the player the creature was
// declared against, captured from the event. The engine has no
// planeswalker defenders, so "the player or planeswalker" is always
// the player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1734e777-d34a-4526-ae24-ef034f147cc5",
		Name:         "Raid Bombardment",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33SmallCreatureYouControlAttacked(ev, source, g)
			},
			Key: "Raid Bombardment — deal 1 damage to the player that creature is attacking",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Raid Bombardment — deal 1 damage to the player that creature is attacking", nil)
				item.Params.Player = ev.Target
				return item
			},
			Effect: b33DamageDefendingPlayerFromSource,
		}},
	})
}
