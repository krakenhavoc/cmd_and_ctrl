package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Beckoning — {B} Sorcery:
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 Return target creature card from your graveyard to your hand. If
//	 this spell was kicked, instead return two target creature cards
//	 from your graveyard to your hand."
//
// The kicked clause is the same predicate with a count of two (#1716,
// WhenPaid) — Peerless Recycling's gift shape on a kicker. A kicked
// cast must name two cards; one that left the graveyard in response is
// skipped and the other still returns (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67a48e3f-2388-42a9-a8b1-97c08761f807",
		Name:         "Blood Beckoning",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{3}"),
			effects.TargetCardInGraveyard("two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(2, 2))},
		Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
