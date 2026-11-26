package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Onyx Goblet — Artifact {2}{B} (Shards of Alara, S82 OLD batch 512, #5588):
//
//	"{T}: Target player loses 1 life."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e070d7a-6a1e-4cce-b196-73f4052c049c",
		Name:         "Onyx Goblet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target player loses 1 life.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82OldB512TargetPlayerLosesOne,
		}},
	})
}
