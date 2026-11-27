package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra's Liturgy — Enchantment {2}{W}{W} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	{W}, Sacrifice this enchantment: Destroy up to X target artifacts and/or enchantments, where X is the number of verse counters on this enchantment.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is fixed as the ability is announced, while it is still on the battlefield (CR 601.2b, 602.2b; TargetSpec.MaxFrom), and each target is judged again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17b7b7e4-4ae8-495d-bf0c-d18697f250ea",
		Name:         "Serra's Liturgy",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Serra's Liturgy")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice this enchantment: Destroy up to X target artifacts and/or enchantments, where X is the number of verse counters on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.S82OldB724SacrificeCost("{W}"),
			Targets: effects.S82OldB724UpToXTargets(effects.TargetPermanent("up to X target artifacts and/or enchantments, where X is the number of verse counters on this enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
			Effect: effects.S82OldB724ForEachLegalTarget(func(ctx *effects.Context, t game.TargetRef) error {
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}),
		}},
	})
}
