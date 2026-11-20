package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emissary of Despair — Creature — Spirit {1}{B}{B} (Darksteel, #5605):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 loses 1 life for each artifact they control."
//
// Hoard Hauler's count of the damaged player's artifacts, read as the
// trigger resolves (CR 608.2h); with none it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4efa1654-c822-4477-ad24-f379ed7ecb11",
		Name:            "Emissary of Despair",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Emissary of Despair — that player loses 1 life for each artifact they control",
				func(g *game.Game, item *game.StackItem) error {
					victim := item.Trigger.Event.Target
					n := len(permanentsControlledByMatching(g, victim, Artifact()))
					if n <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -n)
				}),
		},
	})
}
