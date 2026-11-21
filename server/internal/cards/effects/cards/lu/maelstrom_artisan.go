package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maelstrom Artisan // Rocket Volley — Creature — Minotaur Sorcerer
// {1}{R}{R}, 3/2 // Sorcery {1}{R} (Secrets of Strixhaven, #4456;
// preparation card, CR 722):
//
//	"Haste
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Rocket Volley — "Destroy target nonbasic land."
//
// "Enters prepared" is SelfEntersPrepared. "Nonbasic" is the supertype
// (CR 205.4c, NonbasicLand): a shock land is a Mountain and still a legal
// target, and a Snow-Covered Swamp is basic and not one.
//
// No simplification.
const MaelstromArtisanOracleID = "03931270-504a-4b0b-af3f-918bf124ea07"

func init() {
	effects.Register(effects.Spec{
		OracleID:        MaelstromArtisanOracleID,
		Name:            "Maelstrom Artisan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     MaelstromArtisanOracleID + "#1",
		Name:         "Rocket Volley",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonbasic land", effects.NonbasicLand()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
