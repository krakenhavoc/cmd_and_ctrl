package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shambleshark — Creature — Shark Crab {G}{U}, 2/1 (Gatecrash, #5565):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Evolve (Whenever a creature you control enters, if that creature has
//	 greater power or toughness than this creature, put a +1/+1 counter on
//	 this creature.)"
//
// Evolve is the engine's keyword trigger (game/evolve.go, #1805);
// the other keyword is the engine's too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0e235b96-9507-4f4b-9578-4fb8ed854cc2",
		Name:            "Shambleshark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", game.KeywordEvolve},
	})
}
