package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farseek — Sorcery {1}{G}:
//
//	"Search your library for a Plains, Island, Swamp, or Mountain
//	card, put it onto the battlefield tapped, then shuffle."
//
// Note the printed text says card TYPES, not "basic" — in paper
// Farseek fetches a Hallowed Fountain or a Bayou, which is most of
// why it's played over Rampant Growth in three-plus colours.
//
// A card in a library has no layer cache, so HasSubtype reads
// straight off the printed type line (game.Card.HasSubtype /
// game.ParseTypeLine) — the same read the fetchland cycle's
// landWithEitherSubtype already relies on to find a Watery Grave.
// That means the real clause is expressible without half-measures:
// IsLandWithAnySubtype composes IsLandWithSubtype (helpers.go) over
// the four named types, so a Godless Shrine (Plains Swamp), an
// Overgrown Tomb (Swamp Forest), a Temple Garden (Forest Plains) and
// an Indatha Triome (Plains Swamp Forest) are all legal finds, and a
// Wastes — which prints none of the four land types — is correctly
// not.
func init() {
	Register(Spec{
		OracleID:     "495e52e6-4c2b-4574-9474-eadbdcc8b4ac",
		Name:         "Farseek",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsLandWithAnySubtype("Plains", "Island", "Swamp", "Mountain"),
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Farseek — a Plains, Island, Swamp or Mountain",
			}.Apply(ctx)
		},
	})
}
