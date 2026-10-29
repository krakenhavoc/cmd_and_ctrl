package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreaded Bat-Cloud — Creature — Bat {4}{B}, 4/2 (The Hobbit):
//
//	"This spell costs {3} less to cast if a creature died this turn.
//	 Flying, deathtouch"
//
// The condition is the table-wide tally of creatures that died this turn
// (b11CreaturesDiedThisTurn, Barad-dûr's "a creature died this turn" read at
// cast time): any creature, whoever's, token or card. The reduction spends
// against generic mana only (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3cc64ce6-45e2-4dc3-9b28-d7206a131873",
		Name:            "Dreaded Bat-Cloud",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if a creature died this turn.",
				func(q game.CostQuery) bool { return q.Game != nil && b11CreaturesDiedThisTurn(q.Game) > 0 }),
		},
	})
}
