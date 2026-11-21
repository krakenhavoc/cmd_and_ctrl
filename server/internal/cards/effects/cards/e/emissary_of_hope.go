package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emissary of Hope — Creature — Spirit {1}{W}{W} (Darksteel, #5605):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you gain 1
//	 life for each artifact that player controls."
//
// Hoard Hauler's count of the damaged player's artifacts, read as the
// trigger resolves (CR 608.2h); the Emissary's controller gains that
// much.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "62a6dbdb-70fc-43ab-8e63-367a609ecdeb",
		Name:            "Emissary of Hope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Emissary of Hope — you gain 1 life for each artifact that player controls",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.PermanentsControlledByMatching(g, item.Trigger.Event.Target, effects.Artifact()))
					return effects.GainLife{Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
