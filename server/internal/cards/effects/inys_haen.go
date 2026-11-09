package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inys Haen — Plane — Cridhe (Commander Masters, #4870):
//
//	"When you planeswalk to Inys Haen and at the beginning of your
//	 upkeep, mill three cards.
//	 When you planeswalk away from Inys Haen, each player returns all land
//	 cards from their graveyard to the battlefield tapped.
//	 Whenever chaos ensues, return target nonland card from your
//	 graveyard to your hand."
//
// A plane (ADR 0216 §6). The first ability is one ability with two trigger
// conditions that mills three cards from the planar controller's library.
// The planeswalk-away ability (CR 901.11b) is the plane's own, looking back
// at it as it leaves: every player's land cards return together, each to
// the battlefield tapped under its owner's control. The chaos ability (CR
// 311.7) returns a nonland card from the planar controller's graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61c5f2cb-f39c-487c-9786-a8dad4663f4f",
		Name:         "Inys Haen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Inys Haen — mill three cards", Do(MillCards{N: 3})),
			WhenYouPlaneswalkAwayFromThis("Inys Haen — each player returns all land cards from their graveyard to the battlefield tapped",
				Do(Instruction(func(ctx *Context) error {
					var ids []uuid.UUID
					for _, p := range ctx.Game.Seats {
						if p != nil {
							ids = append(ids, graveyardCardIDs(ctx, p.ID, game.Card.IsLand)...)
						}
					}
					return ReturnFromGraveyardTogether{Targets: ids, Tapped: true}.Apply(ctx)
				}))),
			Targeting(WheneverChaosEnsues("Inys Haen — return target nonland card from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target nonland card from your graveyard", Nonland(), YouOwn())),
		},
	})
}
