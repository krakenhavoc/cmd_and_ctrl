package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exalted Dragon — Creature — Dragon {4}{W}{W}, 5/5 (Exodus, S82 OLD batch
// 745, #5626):
//
//	"Flying
//	 This creature can't attack unless you sacrifice a land. (This cost is
//	 paid as attackers are declared.)"
//
// An attack cost owed by this creature alone (#6190, CR 508.1h): the land is
// sacrificed as the declaration is made, chosen by its controller, and a
// declaration that does not pay it is refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c901805e-c196-4b9d-8d5a-e17367fb37bc",
		Name:            "Exalted Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AttackTaxes: []game.AttackTax{{
			Label:      "This creature can't attack unless you sacrifice a land.",
			OnAttacker: true,
			AppliesTo:  func(q game.AttackTaxQuery) bool { return q.Attacker.InstanceID == q.Source.InstanceID },
			NonMana:    game.CombatCost{Sacrifice: effects.SacrificeSpec("a land", effects.Land()), Label: "a land"},
		}},
	})
}
