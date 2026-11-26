package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Polis Crusher — Creature — Cyclops {2}{R}{G}, 4/4 (Theros):
//
//	"Trample, protection from enchantments
//	 {4}{R}{G}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)
//	 Whenever this creature deals combat damage to a player, if this creature is
//	 monstrous, destroy target enchantment that player controls."
//
// Monstrosity is the shared activated ability. The trigger is an intervening
// "if" (CR 603.4) on the Crusher being monstrous, checked as it triggers and
// again as it resolves; the enchantment is chosen as the trigger goes on the
// stack, among those the damaged player controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ffd6f55f-2592-46a5-94da-cb9424783004",
		Name:            "Polis Crusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "protection from enchantments"},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{4}{R}{G}"), 3)},
		Triggered: []game.TriggeredAbility{
			effects.OldB396PolisCrusherTrigger(),
		},
	})
}
