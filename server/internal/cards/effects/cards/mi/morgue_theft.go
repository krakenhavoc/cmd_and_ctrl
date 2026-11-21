package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morgue Theft — Sorcery {1}{B} (Odyssey, #5613):
//
//	"Return target creature card from your graveyard to your hand.
//	 Flashback {4}{B} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Raise Dead with flashback: the shared returnTargetFromGraveyardToHand
// body, and a card that left the graveyard in response is no longer a
// legal target, so the spell does nothing (CR 608.2b). Cast with
// flashback, Morgue Theft is on the stack, not in the graveyard, so it
// can't target itself, and it is exiled as it leaves the stack (CR
// 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0b8a508b-20c0-4a87-a73a-08af72c39a0a",
		Name:             "Morgue Theft",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{B}")},
		OnResolve:        effects.ReturnTargetFromGraveyardToHand,
	})
}
