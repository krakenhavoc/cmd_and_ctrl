package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scotty, Chief Engineer — {3}{U} Legendary Creature — Human Artificer
// 3/4 (Star Trek Commander, #4974):
//
//	"Tap an untapped artifact you control: Target creature can't be
//	 blocked this turn."
//
// The cost is TapAnUntapped over artifacts (Urza, Lord High Artificer's
// cost on a non-mana ability): not the {T} symbol, so an artifact that
// arrived this turn may pay it (CR 302.6), and it does not target. Scotty
// is not an artifact, so he never pays it himself. The effect is Zhuge
// Jin's targetCantBeBlockedThisTurn; a target gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bd553db-2c7d-42cd-8ded-ab2c955b8f4e",
		Name:         "Scotty, Chief Engineer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped artifact you control: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.TapAnUntapped("an untapped artifact you control", effects.Artifact()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeBlockedThisTurn("Scotty, Chief Engineer"),
		}},
	})
}
