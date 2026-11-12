package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disturbed Burial — Sorcery {1}{B} (Tempest, #5628):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Return target creature card from your graveyard to your hand."
//
// Raise Dead's target clause and the shared
// returnTargetFromGraveyardToHand body, with Capsize's buyback
// (CR 702.27a). With no creature card in your graveyard there is no
// legal target and the spell can't be cast; a card that left the
// graveyard in response means the spell doesn't resolve (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "cdbaa21d-5896-4b8b-9a08-f0839ccfddfc",
		Name:          "Disturbed Burial",
		Completeness:  CompletenessFull,
		Targets:       TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve:     returnTargetFromGraveyardToHand,
	})
}
