package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brokkos, Apex of Forever — Legendary Creature — Nightmare Beast Elemental
// {2}{B}{G}{U}, 6/6 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {2}{U/B}{G}{G} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into the
//	 creature on top plus all abilities from under it.)
//	 Trample
//	 You may cast this card from your graveyard using its mutate ability."
//
// Mutate is the engine's (ADR 0223), offered twice with the shape Blitz uses
// for "cast this card from your graveyard using its blitz ability"
// (MutateFromHandOrGraveyard): "mutate" is bound to the hand and
// "mutate-graveyard" to the graveyard, and the card lists the graveyard in
// CastableZones. The graveyard opens only through its own offer, so Brokkos
// cannot be cast from there for its mana cost, and from there it is cast over
// or under a non-Human creature you own like any mutating spell (CR 702.140a),
// falling back to an ordinary creature spell if that creature is gone as it
// resolves (CR 702.140b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e9137b54-3464-4f57-95e3-f94bf2ac854e",
		Name:             "Brokkos, Apex of Forever",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: effects.MutateFromHandOrGraveyard("{2}{U/B}{G}{G}"),
	})
}
