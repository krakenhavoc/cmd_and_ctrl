package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "62a6dbdb-70fc-43ab-8e63-367a609ecdeb",
		Name:            "Emissary of Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Emissary of Hope — you gain 1 life for each artifact that player controls",
				func(g *game.Game, item *game.StackItem) error {
					n := len(permanentsControlledByMatching(g, item.Trigger.Event.Target, Artifact()))
					return GainLife{Amount: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
