package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solve for Disappointment — Sorcery {1}{B} (Reality Fracture):
//
//	"Target opponent reveals their hand. You choose a nonland
//	 permanent card from it. That player discards that card.
//	 Empower Jace 1."
//
// ADR 0139 proof card: the revealed-hand pick (ADR 0116), then the
// keyword action. The pick pauses the resolution, so Empower Jace is
// the pick's continuation and happens after the discard, in printed
// order. A hand with no nonland permanent card is revealed and nothing
// is discarded (CR 609.3); Jace is empowered either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3241a3f5-7ae5-4497-9dd2-a55f3370a055",
		Name:         "Solve for Disappointment",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return EmpowerJace{N: 1}.Apply(ctx)
			}
			return ChooseFromRevealedHand{
				Player: victim,
				Filter: NonlandPermanentCard(),
				Label:  "nonland permanent card",
				Then:   solveForDisappointmentEmpower,
			}.Apply(ctx)
		},
	})
}

// solveForDisappointmentEmpower is the card's "Empower Jace 1.", run
// once the chosen card is discarded (or nothing could be chosen).
var solveForDisappointmentEmpower = RevealedPickThen("revealed-pick/solve-for-disappointment-empower-jace",
	func(ctx *Context, _ game.RevealedPick) error {
		return EmpowerJace{N: 1}.Apply(ctx)
	})
