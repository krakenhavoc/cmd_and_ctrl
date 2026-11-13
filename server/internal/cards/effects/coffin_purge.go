package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coffin Purge — Instant {B} (Odyssey, #5613):
//
//	"Exile target card from a graveyard.
//	 Flashback {B} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Any graveyard's card, the shared exileTargetCardsOnResolve body; a card
// that left its graveyard in response is no longer a legal target, and
// the spell does nothing (CR 608.2b). Cast with flashback, Coffin Purge
// is on the stack, so it can't target itself.
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9db5aa54-3deb-45ee-82a2-e88a4367c6d9",
		Name:             "Coffin Purge",
		Completeness:     CompletenessFull,
		Targets:          TargetCardInGraveyard("target card from a graveyard"),
		Purpose:          ForTargets(ExilesFromGraveyard(0, "")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{B}")},
		OnResolve:        exileTargetCardsOnResolve,
	})
}
