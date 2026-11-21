package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forgotten Sentinel — Artifact Creature — Golem {4}, 4/3 (Core Set 2021, #5511):
//
//	"This creature enters tapped."
//
// Diregraf Ghoul's enters-tapped replacement, and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bd4a0bd-d10c-434a-8254-89f38740aafd",
		Name:         "Forgotten Sentinel",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
