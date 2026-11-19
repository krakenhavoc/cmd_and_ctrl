package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Refurbished Familiar — Artifact Creature — Zombie Rat {3}{B}, 2/1
// (Modern Horizons 3):
//
//	"Affinity for artifacts
//	 Flying
//	 When this creature enters, each opponent discards a card. For each
//	 opponent who can't, you draw a card."
//
// The discard-or-draw body is Aclazotz's, shared.
func init() {
	Register(Spec{
		OracleID:        "009e1638-9055-4872-93e1-85841bce4648",
		Name:            "Refurbished Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Refurbished Familiar — each opponent discards a card; for each who can't, you draw a card", aclazotzAttack),
		},
	})
}
