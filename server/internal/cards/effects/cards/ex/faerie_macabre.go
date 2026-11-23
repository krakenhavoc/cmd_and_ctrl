package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Macabre — Creature — Faerie Rogue {1}{B}{B}, 2/2:
//
//	"Flying
//	 Discard this card: Exile up to two target cards from graveyards."
//
// A hand ability (CR 113.6, ADR 0062) paid by discarding itself, Boseiju's
// channel shape; the targets are up to two cards in any graveyards (no
// single-graveyard rule), exiled together as it resolves with ExileTargetCards,
// which skips a target that left its graveyard in response (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7d8beb6-472c-4758-9876-e8cbf98f4f72",
		Name:            "Faerie Macabre",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard this card: Exile up to two target cards from graveyards",
			Cost:    game.AbilityCost{DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Targets: effects.TargetCardInGraveyard("up to two target cards from graveyards").WithCount(0, 2),
			Effect:  effects.ExileTargetCards,
		}},
	})
}
