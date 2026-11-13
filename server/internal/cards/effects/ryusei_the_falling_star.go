package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ryusei, the Falling Star — Legendary Creature — Dragon Spirit {5}{R},
// 5/5 (Champions of Kamigawa, #5603):
//
//	"Flying
//	 When Ryusei dies, it deals 5 damage to each creature without flying."
//
// Seismic Rupture's sweep at 5 on a dies trigger (CR 700.4): the
// creatures without flying are read post-layer as it resolves, so one
// granted flying in response is spared, and Ryusei, gone, deals the
// damage as it last existed. No player is dealt damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5e99ba3b-dc2b-4c3c-84ce-228d4772cfee",
		Name:            "Ryusei, the Falling Star",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			// ADR 0126 §6: a creature with flying is spared.
			TriggerWithPurpose(WhenThisDies("Ryusei, the Falling Star — 5 damage to each creature without flying",
				func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item), And(Creature(), WithoutKeyword("flying")), 5)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 5, Partial: true}}),
		},
	})
}
