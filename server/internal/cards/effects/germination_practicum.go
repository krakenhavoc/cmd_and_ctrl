package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Germination Practicum — Sorcery — Lesson {3}{G}{G} (Secrets of
// Strixhaven, #4456):
//
//	"Put two +1/+1 counters on each creature you control.
//	 Paradigm (Then exile this spell. After you first resolve a spell
//	 with this name, you may cast a copy of it from exile without paying
//	 its mana cost at the beginning of each of your first main phases.)"
//
// The counters are one placement of two per creature
// (b11PutCountersOnEachCreatureYouControl), so Hardened Scales adds one
// to each, not two. Paradigm is the printed keyword and the engine does
// the rest (game/paradigm.go, CR 702.192): the spell exiles itself as it
// resolves, and from then on each of your first main phases offers a free
// cast of a copy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8d22139-7b25-4bb1-b0d9-d4c622401edf",
		Name:            "Germination Practicum",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordParadigm},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b11PutCountersOnEachCreatureYouControl(ctx.Game, item, 2)
		},
	})
}
