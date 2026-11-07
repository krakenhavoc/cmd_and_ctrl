package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Furtive Courier — Creature — Merfolk Advisor {2}{U}, 3/2 (Murders at
// Karlov Manor, #4469):
//
//	"This creature can't be blocked as long as you've sacrificed an
//	 artifact this turn.
//	 Whenever this creature attacks, draw a card, then discard a card."
//
// The evasion is a block rule on the creature itself (CantBeBlockedWhile,
// CR 509.1b), checked as blockers are declared against the turn's
// card-type sacrifice tally (SacrificedAnArtifactThisTurn, #5224): an
// artifact sacrificed before blockers, even during combat, makes it
// unblockable, and a token counts. The attack trigger is the shared loot
// body (b16DrawThenDiscard): the discard is chosen after the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92d24b75-419b-4068-89c1-db9945f74972",
		Name:         "Furtive Courier",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedWhile(OnSelf(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
				return SacrificedAnArtifactThisTurn(g, you)
			}),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Furtive Courier — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) }),
		},
	})
}
