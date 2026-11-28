package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torch Song — Enchantment {2}{R} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	{2}{R}, Sacrifice this enchantment: It deals X damage to any target, where X is the number of verse counters on this enchantment.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is the verse counters it had as it last existed on the battlefield (CR 608.2h). The damage is dealt by the sacrificed enchantment as it last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69826a0d-3ec9-47e7-b129-2fc13b41a07e",
		Name:         "Torch Song",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Torch Song")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this enchantment: It deals X damage to any target, where X is the number of verse counters on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.S82OldB724SacrificeCost("{2}{R}"),
			Targets: effects.TargetAny(),
			Effect: effects.S82OldB724ForEachLegalTarget(func(ctx *effects.Context, t game.TargetRef) error {
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: effects.S82OldB724X(ctx)}.Apply(ctx)
			}),
		}},
	})
}
