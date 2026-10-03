package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rock Jockey — Creature — Goblin {2}{R}:
//
//	"You can't cast Rock Jockey if you've played a land this turn.
//	 You can't play lands if this creature was cast this turn."
//
// ADR 0109 §4 (#1895). Two halves of one lock:
//
//   - The first sentence is the spell's own "cast only if" condition
//     (CR 101.2, the CastCondition slot Legendary sorceries use): read
//     off the engine's per-turn land-play tally, which counts lands
//     played as special actions and during resolutions alike.
//   - The second is a LandPlayRestriction on the permanent, read off its
//     CONTROLLER: "you" is the Jockey's controller, and a Jockey that
//     was cast this turn and has since changed hands restricts its new
//     controller. "Was cast this turn" is the turn's event log (see
//     permanentWasCastThisTurn): a Jockey that was bounced and put back
//     is a new object that was not cast (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09fff6a3-6a54-4101-a587-55d7113b9639",
		Name:         "Rock Jockey",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return g.LandsPlayedThisTurnFor(controller) == 0
		},
		CastConditionLabel: "You can't cast Rock Jockey if you've played a land this turn.",
		LandPlayRestrictions: []game.LandPlayRestriction{{
			Label: "You can't play lands if this creature was cast this turn.",
			Forbids: func(q game.LandPlayQuery) bool {
				return q.Source.Controller == q.Player && permanentWasCastThisTurn(q.Game, q.Source.InstanceID)
			},
		}},
	})
}
