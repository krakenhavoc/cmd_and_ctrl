package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "010ec11a-0484-486d-b7c0-de65593bf457",
		Name:         "Sanguine Indulgence",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if you've gained 3 or more life this turn.",
				func(q game.CostQuery) bool { return q.Game != nil && b15LifeGainedThisTurn(q.Game, q.Controller) >= 3 }),
		},
		Targets:   TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
		OnResolve: ResolveSteps(Instruction(returnLegalGraveyardTargetsToHand)),
	})
}
