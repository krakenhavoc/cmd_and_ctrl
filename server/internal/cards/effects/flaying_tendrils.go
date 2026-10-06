package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flaying Tendrils — Sorcery {1}{B}{B}, devoid:
//
//	"Devoid (This card has no color.)
//	 All creatures get -2/-2 until end of turn. If a creature would die
//	 this turn, exile it instead."
//
// Devoid is declared in PrintedKeywords, and the engine reads it as CR
// 702.114a's colour-defining ability (#2152), so the card is colourless
// in every zone. "If a
// creature would die this turn" changes no characteristic, so its set is
// read as each creature would die (CR 611.2c, ADR 0108 §1): a creature
// cast after Flaying Tendrils resolved is exiled too if it dies this
// turn. The -2/-2 locks its set as it begins.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "cc2d016a-af44-427b-a25a-593274369449",
		Name:            "Flaying Tendrils",
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		OnResolve:       allCreaturesShrinkThenExileIfTheyDie(-2, false),
	})
}
