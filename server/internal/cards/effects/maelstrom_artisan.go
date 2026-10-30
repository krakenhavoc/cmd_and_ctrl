package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const maelstromArtisanOracleID = "03931270-504a-4b0b-af3f-918bf124ea07"

func init() {
	Register(Spec{
		OracleID:        maelstromArtisanOracleID,
		Name:            "Maelstrom Artisan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     maelstromArtisanOracleID + "#1",
		Name:         "Rocket Volley",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonbasic land", NonbasicLand()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
