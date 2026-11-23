package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Buried in the Garden — Enchantment — Aura {2}{G}{W} (Murders at Karlov
// Manor, #4469):
//
//	"Enchant land
//	 When this Aura enters, exile target nonland permanent you don't
//	 control until this Aura leaves the battlefield.
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional one mana of any color."
//
// Makeshift Binding's "until this leaves" exile (CR 610.3: the return is
// a one-shot made as the Aura leaves, before anyone gets priority, and
// the permanent comes back under its owner's control) on Fertile
// Ground's mana trigger. If the Aura leaves before its enters trigger
// resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1f0eab3-681b-4c2d-b590-555951920880",
		Name:         "Buried in the Garden",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Buried in the Garden — exile target nonland permanent you don't control until this Aura leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Buried in the Garden — the exiled permanent returns when Buried in the Garden leaves the battlefield")),
				effects.TargetPermanent("target nonland permanent you don't control", effects.Nonland(), effects.OpponentControls())),
		},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverAttachedTapsForMana(
				"Buried in the Garden — add an additional one mana of any color",
				"{W|U|B|R|G}"),
		},
	})
}
