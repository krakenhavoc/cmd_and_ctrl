package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reverence — Enchantment {2}{W}{W} (Saviors of Kamigawa, #5601):
//
//	"Creatures with power 2 or less can't attack you."
//
// Queen Mother Ramonda's attack restriction without its condition: it keeps
// those creatures from attacking Reverence's controller, and not their
// planeswalkers, and is judged at declaration on each attacker's effective
// power then (CR 508.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60963292-2dee-4b1c-9c1c-c35cf8ea3496",
		Name:         "Reverence",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.CreaturesMatchingCantAttackYou("Creatures with power 2 or less can't attack you.",
			game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2))},
	})
}
