package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Purple Worm — Creature — Worm {5}{G}{G}, 8/7 (Adventures in the Forgotten
// Realms, S82):
//
//	"This spell costs {2} less to cast if a creature died this turn.
//	 Ward {2}"
//
// The discount is Avenge's self cost modifier over the table-wide tally of
// creatures that died (any controller, tokens included), read as the spell
// is cast. Ward is the shared ward trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "179cc807-5150-4513-88b2-73cb05009b7d",
		Name:         "Purple Worm",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if a creature died this turn.", ACreatureDiedThisTurnSpell()),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Purple Worm — ward {2}"),
		},
	})
}
