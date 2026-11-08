package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f1f0eab3-681b-4c2d-b590-555951920880",
		Name:         "Buried in the Garden",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Buried in the Garden — exile target nonland permanent you don't control until this Aura leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Buried in the Garden — the exiled permanent returns when Buried in the Garden leaves the battlefield")),
				TargetPermanent("target nonland permanent you don't control", Nonland(), OpponentControls())),
		},
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForMana(
				"Buried in the Garden — add an additional one mana of any color",
				"{W|U|B|R|G}"),
		},
	})
}
