package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Air Servant — Creature — Elemental {4}{U}, 4/3 (Magic 2011, #5579):
//
//	"Flying
//	 {2}{U}: Tap target creature with flying."
//
// Flying is the printed keyword. The ability is Storm Elemental's tap row
// at a mana cost: any creature with flying, its own included, and a target
// that lost flying in response is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "49638d96-e83e-4aab-bafb-d6183a6e3ff6",
		Name:            "Air Servant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Tap target creature with flying.",
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
