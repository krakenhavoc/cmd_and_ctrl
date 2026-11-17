package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tinker — Sorcery {2}{U} (Urza's Legacy, #5623):
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Search your library for an artifact card, put that card onto the
//	 battlefield, then shuffle."
//
// Crop Rotation's shape for artifacts. The sacrifice is an additional cost
// (CR 601.2f) paid while casting (CR 601.2h), so a countered Tinker does not
// hand the artifact back, and the sacrificed artifact can't be the one
// found. Any artifact card, artifact creatures and artifact lands included,
// enters untapped; its own enters-tapped replacement still runs. The search
// shuffles even when it finds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "254878f0-be90-4653-a395-0c41258fceaf",
		Name:           "Tinker",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact", Artifact()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsArtifact() },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Tinker — an artifact card",
			}.Apply(ctx)
		},
	})
}
