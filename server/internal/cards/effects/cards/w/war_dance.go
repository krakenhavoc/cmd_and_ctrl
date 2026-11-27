package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War Dance — Enchantment {G} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	Sacrifice this enchantment: Target creature gets +X/+X until end of turn, where X is the number of verse counters on this enchantment.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is the verse counters it had as it last existed on the battlefield (CR 608.2h). There is no mana cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecfc81df-1271-47ca-9184-49bf5798e195",
		Name:         "War Dance",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("War Dance")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Target creature gets +X/+X until end of turn, where X is the number of verse counters on this enchantment.",
			Purpose: effects.S82OldB724PumpPurpose(),
			Cost:    effects.S82OldB724SacrificeCost(""),
			Targets: effects.TargetCreature("target creature"),
			Effect: effects.S82OldB724ForEachLegalTarget(func(ctx *effects.Context, t game.TargetRef) error {
				x := effects.S82OldB724X(ctx)
				return effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "War Dance — +X/+X"}.Apply(ctx)
			}),
		}},
	})
}
