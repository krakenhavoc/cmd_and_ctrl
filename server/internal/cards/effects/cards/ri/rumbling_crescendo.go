package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rumbling Crescendo — Enchantment {3}{R}{R} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	{R}, Sacrifice this enchantment: Destroy up to X target lands, where X is the number of verse counters on this enchantment.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is fixed as the ability is announced, while it is still on the battlefield (CR 601.2b, 602.2b; TargetSpec.MaxFrom), and each target is judged again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e7e19c2-5399-433d-9653-d956f11838f5",
		Name:         "Rumbling Crescendo",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Rumbling Crescendo")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this enchantment: Destroy up to X target lands, where X is the number of verse counters on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.S82OldB724SacrificeCost("{R}"),
			Targets: effects.S82OldB724UpToXTargets(effects.TargetPermanent("up to X target lands, where X is the number of verse counters on this enchantment", effects.Land())),
			Effect: effects.S82OldB724ForEachLegalTarget(func(ctx *effects.Context, t game.TargetRef) error {
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}),
		}},
	})
}
