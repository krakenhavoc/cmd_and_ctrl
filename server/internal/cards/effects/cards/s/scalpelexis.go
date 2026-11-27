package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scalpelexis — Creature — Beast {4}{U}, 1/5 (Judgment, #5611):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 exiles the top four cards of their library. If two or more of those
//	 cards have the same name, repeat this process."
//
// The exile is the keyed top-of-library exile; its tail reads the names of the
// four cards that landed and starts the process again while two share one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99eb50ef-352f-47c4-91e3-32813cbe0649",
		Name:            "Scalpelexis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Scalpelexis — that player exiles the top four cards of their library, and repeats if two share a name",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.OldB623ScalpelexisExile(effects.NewContext(g, item), item.Trigger.Event.Target)
				}),
		},
	})
}
