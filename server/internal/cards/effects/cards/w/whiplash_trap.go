package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whiplash Trap — Instant — Trap {3}{U}{U} (Planechase, #5584):
//
//	"If an opponent had two or more creatures enter the battlefield under
//	 their control this turn, you may pay {U} rather than pay this spell's
//	 mana cost.
//	 Return two target creatures to their owners' hands."
//
// The trap clause is an alternative cost (CR 118.9) with Mindbreak Trap's
// shape. Its condition reads each opponent's entry tally for the turn
// (Game.EnteredWithCardTypeThisTurn, "creature"), recorded as each
// permanent entered with its types at that moment: tokens count, a
// creature that has since left still counts, and two opponents with one
// creature each do not add up to two.
//
// The two creatures go back in one return (Specimen Freighter's), so a
// returned commander may stop for its owner's command-zone answer
// (CR 903.9); a target that left in response is skipped and the other
// still returns (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bd26692-c030-4493-883e-2d26b3a1d025",
		Name:         "Whiplash Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent had two or more creatures enter the battlefield under their control this turn, pay {U}",
			ManaCost:  "{U}",
			Condition: AnOpponentHadTwoOrMoreCreaturesEnterThisTurn,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.BounceCardsToHandForEffect(effects.LegalTargetCardIDs(ctx))
			return nil
		},
	})
}

// anOpponentHadTwoOrMoreCreaturesEnterThisTurn is Whiplash Trap's
// condition: one opponent, on their own, had two or more creatures enter
// the battlefield under their control this turn.
func AnOpponentHadTwoOrMoreCreaturesEnterThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		if g.EnteredWithCardTypeThisTurn(p.ID, "creature") >= 2 {
			return true
		}
	}
	return false
}
