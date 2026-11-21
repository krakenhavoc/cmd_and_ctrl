package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Indulgence — Sorcery {3}{B} (Jumpstart, #4892):
//
//	"This spell costs {3} less to cast if you've gained 3 or more life
//	 this turn.
//	 Return up to two target creature cards from your graveyard to your
//	 hand."
//
// The reduction reads the caster's life gained this turn as the cost is
// determined (CR 601.2f) and reduces only the generic part, so the spell
// costs {B} after 3 life. Life lost again later in the turn does not undo
// a gain: the tally is of life gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "010ec11a-0484-486d-b7c0-de65593bf457",
		Name:         "Sanguine Indulgence",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(3, "This spell costs {3} less to cast if you've gained 3 or more life this turn.",
				func(q game.CostQuery) bool {
					return q.Game != nil && effects.B15LifeGainedThisTurn(q.Game, q.Controller) >= 3
				}),
		},
		Targets:   effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand)),
	})
}
