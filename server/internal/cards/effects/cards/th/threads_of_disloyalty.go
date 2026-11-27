package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Threads of Disloyalty — Enchantment — Aura {1}{U}{U} (Betrayers of Kamigawa, S82 OLD
// b594):
//
//	"Enchant creature with mana value 2 or less
//	 You control enchanted creature."
//
// Control Magic with a mana-value bound on the enchant clause. The clause is the spec the
// state-based action re-runs (CR 704.5m), so an enchanted creature whose mana value rises
// above 2 (a copy effect, a changed cost) takes the Aura to the graveyard and control goes
// home.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d86735c-12e5-457a-beb4-057ab220d18b",
		Name:         "Threads of Disloyalty",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.ManaValueLE(2)),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
	})
}
