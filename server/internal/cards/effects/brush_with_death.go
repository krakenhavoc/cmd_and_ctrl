package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brush with Death — Sorcery {2}{B} (Stronghold, #5627):
//
//	"Buyback {2}{B}{B} (You may pay an additional {2}{B}{B} as you cast
//	 this spell. If you do, put this card into your hand as it resolves.)
//	 Target opponent loses 2 life. You gain 2 life."
//
// Syphon Life's shared targetPlayerLosesLifeYouGain body at 2, on a
// "target opponent" clause, with Capsize's buyback (CR 702.27a). The
// spell's one target is the opponent: if that player has left the game in
// response, the spell doesn't resolve (CR 608.2b), so no life is gained
// and the card goes to the graveyard even with buyback paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "d87e580f-eca1-4f91-962b-82d684cba5e7",
		Name:          "Brush with Death",
		Completeness:  CompletenessFull,
		Targets:       TargetPlayer("target opponent", Opponent()),
		OptionalCosts: []game.AdditionalCost{Buyback("{2}{B}{B}")},
		OnResolve:     targetPlayerLosesLifeYouGain(2),
	})
}
