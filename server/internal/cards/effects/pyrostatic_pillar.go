package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyrostatic Pillar — Enchantment {1}{R} (Scourge):
//
//	"Whenever a player casts a spell with mana value 3 or less, this
//	 enchantment deals 2 damage to that player."
//
// Chalice of the Void's cast watcher, reading the spell's mana value as it
// sits on the stack ({X} counts as the value chosen, CR 202.3e), and every
// player's spells, its controller's included. The damage goes to the
// player who cast the spell, and the Pillar is its source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b5b87da-5a00-449a-97ef-fa1e2c1c4b4b",
		Name:         "Pyrostatic Pillar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && spellManaValueForEffect(g, ev.CardID) <= 3
			}, "Pyrostatic Pillar — 2 damage to that player", scgB1DamageThatPlayer(2)),
		},
	})
}
