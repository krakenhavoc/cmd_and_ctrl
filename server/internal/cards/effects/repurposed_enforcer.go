package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Repurposed Enforcer — Creature — Human Soldier {1}{W}, 3/2 (Reality
// Fracture):
//
//	"Whenever this creature attacks, empower Jace X, where X is the
//	 number of creatures you control."
//
// ADR 0139: "Empower Jace X" with X counted as the trigger resolves
// (CR 608.2h), through EmpowerJace.Count. The Enforcer counts itself
// while it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b775a404-ec0a-47ff-bb9d-eae2273d945b",
		Name:         "Repurposed Enforcer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Repurposed Enforcer — empower Jace X, where X is the number of creatures you control",
				Do(EmpowerJace{Count: func(ctx *Context) int {
					return creaturesControlledBy(ctx.Game, ctx.Controller())
				}})),
		},
	})
}

// creaturesControlledBy is "the number of creatures you control".
// Caller must hold g.mu.
func creaturesControlledBy(g *game.Game, player uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.IsCreature() {
			n++
		}
	}
	return n
}
