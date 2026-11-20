package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eidolon of the Great Revel — Enchantment Creature — Spirit {R}{R}, 2/2 (Journey into Nyx, #5558):
//
//	"Whenever a player casts a spell with mana value 3 or less, this
//	 creature deals 2 damage to that player."
//
// Pyrostatic Pillar's cast watcher on a creature: every player's spell
// with mana value 3 or less, its controller's included, and the damage
// goes to the player who cast it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fd7018e-1c36-4d7f-8f72-22f2f7f60c60",
		Name:         "Eidolon of the Great Revel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && spellManaValueForEffect(g, ev.CardID) <= 3
			}, "Eidolon of the Great Revel — 2 damage to that player", scgB1DamageThatPlayer(2)),
		},
	})
}
