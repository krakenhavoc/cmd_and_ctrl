package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blazing Firesinger // Seething Song — a preparation card (CR 722, ADR 0090),
// Secrets of Strixhaven (#4456):
//
//	Blazing Firesinger {2}{R}, Creature — Dwarf Bard 2/3:
//	"This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//	Seething Song {2}{R}, Instant:
//	"Add {R}{R}{R}{R}{R}."
//
// "Enters prepared" is SelfEntersPrepared; the engine makes, casts and
// unprepares the copy. The prepare spell is Seething Song word for word, so
// it shares that card's body (seethingSongResolve).
//
// No simplification.
func init() {
	const id = "1bf505a7-292e-4642-83b6-8ff41ebe5d51"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Blazing Firesinger",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Seething Song",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SeethingSongResolve,
	})
}
