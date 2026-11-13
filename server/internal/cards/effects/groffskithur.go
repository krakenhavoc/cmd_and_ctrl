package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Groffskithur — Creature — Beast {5}{G}, 3/3 (Mirrodin, #5606):
//
//	"Whenever this creature becomes blocked, you may return target card
//	 named Groffskithur from your graveyard to your hand."
//
// A becomes-blocked trigger (CR 509.1h), once per combat however many
// creatures block it, with its target chosen as it goes on the stack:
// a card named Groffskithur in its controller's graveyard. With no such
// card there is nothing to target and the trigger is removed. The "you
// may" is asked as it resolves, and a target that has left the graveyard
// by then is illegal, so nothing returns (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32a8ca5f-0170-4f5d-97cd-63e2b74fd653",
		Name:         "Groffskithur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				On(game.EventBecomesBlocked, Self,
					"Groffskithur — return target card named Groffskithur from your graveyard to your hand",
					returnTargetCardToHand),
				TargetCardInGraveyard("target card named Groffskithur in your graveyard", YouOwn(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Name == "Groffskithur" })),
				"Groffskithur — return the Groffskithur from your graveyard to your hand?"),
		},
	})
}
