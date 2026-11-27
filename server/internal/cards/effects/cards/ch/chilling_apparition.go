package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chilling Apparition — Creature — Spirit {2}{B}, 1/1 (Prophecy, S82 OLD b670, #5617):
//
//	"{B}: Regenerate this creature.
//	 Whenever this creature deals combat damage to a player, that player
//	 discards a card."
//
// Rakdos Ringleader's regenerate row and Blazing Specter's discard: the damaged
// player chooses the card (DiscardOne), and an empty hand discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13cfc92c-142a-4016-b86f-ace6bda1a69c",
		Name:         "Chilling Apparition",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Chilling Apparition — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, item.Trigger.Event.Target)
				}),
		},
	})
}
