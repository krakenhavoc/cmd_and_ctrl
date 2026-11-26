package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vectis Silencers — Artifact Creature — Human Rogue {2}{U}, 1/2 (Shards of
// Alara, S82 OLD batch 510, #5588):
//
//	"{2}{B}: This creature gains deathtouch until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e132bba0-40c5-4095-99a3-02cad520ca99",
		Name:         "Vectis Silencers",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{2}{B}", "deathtouch")},
	})
}
