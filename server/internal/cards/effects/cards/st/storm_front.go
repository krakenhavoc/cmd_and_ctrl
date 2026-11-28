package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Front — Enchantment {G} (Tempest, S82 OLD b752, #5628):
//
//	"{G}{G}: Tap target creature with flying."
//
// Air Servant's tap row at {G}{G}: any creature with flying, whoever controls
// it, and a target that lost flying in response is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c8044b7-3999-4825-8448-0a1a325f1b48",
		Name:         "Storm Front",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{G}: Tap target creature with flying.",
			Cost:    effects.ManaCost("{G}{G}"),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
