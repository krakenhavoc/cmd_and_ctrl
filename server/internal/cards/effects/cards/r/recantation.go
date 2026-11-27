package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Recantation — Enchantment {3}{U}{U} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	{U}, Sacrifice this enchantment: Return up to X target permanents to their owners' hands, where X is the number of verse counters on this enchantment.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is fixed as the ability is announced, while it is still on the battlefield (CR 601.2b, 602.2b; TargetSpec.MaxFrom), and each target is judged again as it resolves (CR 608.2b). Each is returned to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02269353-2949-4599-b5a3-f20fdb501285",
		Name:         "Recantation",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Recantation")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice this enchantment: Return up to X target permanents to their owners' hands, where X is the number of verse counters on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.S82OldB724SacrificeCost("{U}"),
			Targets: effects.S82OldB724UpToXTargets(effects.TargetPermanent("up to X target permanents, where X is the number of verse counters on this enchantment")),
			Effect: effects.S82OldB724ForEachLegalTarget(func(ctx *effects.Context, t game.TargetRef) error {
				return effects.BounceToHand{Target: t.ID}.Apply(ctx)
			}),
		}},
	})
}
