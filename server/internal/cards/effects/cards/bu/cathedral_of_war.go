package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathedral of War — Land:
//
//	"This land enters tapped.
//	 Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)
//	 {T}: Add {C}."
//
// Exalted is the engine's keyword (game/exalted.go), read off the land
// while it is on the battlefield. The enters-tapped clause is a real
// CR 614 self-replacement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ff647e4-730a-498f-8f2c-5bd64d5a9780",
		Name:            "Cathedral of War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.ColorlessTapRow()},
	})
}
