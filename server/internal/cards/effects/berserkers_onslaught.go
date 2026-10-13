package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Berserkers' Onslaught — Enchantment {3}{R}{R} (EDHREC rank 2382):
//
//	"Attacking creatures you control have double strike."
//
// The aggro deck's five-mana anthem. A layer 6 static gated on
// Card.AttackingTarget (AttackingCreaturesHave), declared
// DependsOnAttackingStatus so the layer cache is rebuilt when an
// attacker is declared, put onto the battlefield attacking, pulled out
// of combat, or combat ends. It used to be a per-attacker trigger
// because that cache did not move on an attack declaration; the
// invalidation arrived with Ohran Frostfang (#1218), after which this is
// the printed static. Covers a token that enters attacking (CR 508.4) as
// well as a declared attacker, for exactly as long as it is attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85d2948c-1a79-418d-80c4-bc1012a4d313",
		Name:         "Berserkers' Onslaught",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			AttackingCreaturesHave("Attacking creatures you control have double strike.", nil, "double strike"),
		},
	})
}
