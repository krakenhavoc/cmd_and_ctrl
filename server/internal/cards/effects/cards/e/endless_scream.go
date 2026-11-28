package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endless Scream — Enchantment — Aura {X}{B} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature
//	 This Aura enters with X scream counters on it.
//	 Enchanted creature gets +1/+0 for each scream counter on this Aura."
//
// Mikaeus's X entry clause (CR 614.1c) with a "scream" counter kind: the counters
// go on the Aura as it enters, off the X announced for the cast, and the pump
// counts the Aura's own counters on every layer pass (layer 7c), so a counter
// removed or doubled changes it at once. Cast for X = 0 it enters with none and
// gives +0/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "cde3e29c-31fd-4dce-a008-283bbe90bfc6",
		Name:                       "Endless Scream",
		Completeness:               effects.CompletenessFull,
		Targets:                    effects.EnchantCreature(),
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters("scream")},
		Purpose:                    game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 0, func(_ *game.Game, source *game.Card) int { return source.Counters["scream"] }),
		},
	})
}
