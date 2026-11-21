package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "009e1638-9055-4872-93e1-85841bce4648",
		Name:            "Refurbished Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Refurbished Familiar — each opponent discards a card; for each who can't, you draw a card", effects.AclazotzAttack),
		},
	})
}
