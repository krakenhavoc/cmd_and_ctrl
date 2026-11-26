package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattered Perception — Sorcery {2}{R}:
//
//	"Discard all the cards in your hand, then draw that many cards.
//	 Flashback {5}{R}"
//
// Tolarian Winds with flashback: the draw reads the discard's count from the
// step record, after any madness prompt the discard asks (CR 608.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9853d789-b048-4cba-a64d-db2694ca9173",
		Name:             "Shattered Perception",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{R}")},
		OnResolve:        effects.OldB441DiscardHandThenDrawThatMany("shattered-perception/discarded"),
	})
}
