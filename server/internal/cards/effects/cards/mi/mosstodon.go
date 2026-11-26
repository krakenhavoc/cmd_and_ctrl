package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mosstodon — Creature — Plant Elephant {4}{G}, 5/3 (Shards of Alara, S82
// OLD batch 510, #5588):
//
//	"{1}: Target creature with power 5 or greater gains trample until end of
//	 turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a41b610e-125d-4921-81f9-6391f9214d12",
		Name:         "Mosstodon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB510TargetGainsRow(effects.ManaCost("{1}"), "{1}", "with power 5 or greater", "trample", effects.PowerGE(5)),
		},
	})
}
