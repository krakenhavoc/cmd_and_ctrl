package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lilting Refrain — Enchantment {1}{U} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	Sacrifice this enchantment: Counter target spell unless its controller pays {X}, where X is the number of verse counters on this enchantment.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is the verse counters it had as it last existed on the battlefield (CR 608.2h). There is no mana cost. With no verse counters there is nothing to pay and nothing to counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f91a71b7-9bc0-4609-a312-ffede52dc0ad",
		Name:         "Lilting Refrain",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Lilting Refrain")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Counter target spell unless its controller pays {X}, where X is the number of verse counters on this enchantment.",
			Cost:    effects.S82OldB724SacrificeCost(""),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.S82OldB724CounterUnlessXPaid,
		}},
	})
}
