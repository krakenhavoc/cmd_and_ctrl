package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pigment Wrangler // Striking Palette — Creature — Orc Sorcerer {4}{R},
// 4/4 // Sorcery {R} (Secrets of Strixhaven, #4456; preparation card,
// CR 722):
//
//	"Flying
//	 This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//
//	Striking Palette — "When you next cast an instant or sorcery spell this
//	 turn, copy that spell. You may choose new targets for the copy."
//
// "Enters prepared" is SelfEntersPrepared. Striking Palette is Doublecast's
// delayed trigger (WhenYouNextCast over a cast filter of instants and
// sorceries): it waits for a cast, fires on the first one that matches
// (CR 603.7b) and ends with the turn. The copy is put on the stack by a
// triggered ability above the spell, and "choose new targets" is CR 707.10c.
// The Palette itself is a spell that has already been cast, so it can't be
// its own "next" one.
//
// No simplification.
const PigmentWranglerOracleID = "c008d60a-ee06-4411-b8e7-c2df13616105"

func init() {
	effects.Register(effects.Spec{
		OracleID:        PigmentWranglerOracleID,
		Name:            "Pigment Wrangler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     PigmentWranglerOracleID + "#1",
		Name:         "Striking Palette",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.WhenYouNextCast(
				"Striking Palette — copy that spell",
				game.CastFilter{Types: []string{"Instant", "Sorcery"}},
				effects.CopyTheSpellBody,
			).Apply(ctx)
		},
	})
}
