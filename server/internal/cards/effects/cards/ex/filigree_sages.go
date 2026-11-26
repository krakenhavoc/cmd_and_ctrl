package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Filigree Sages — Artifact Creature — Vedalken Wizard {3}{U}, 2/3 (Shards of
// Alara, S82 OLD batch 510, #5588):
//
//	"{2}{U}: Untap target artifact."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6e78c6a-7555-4818-8054-1cdc0fca92e6",
		Name:         "Filigree Sages",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Untap target artifact.",
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
