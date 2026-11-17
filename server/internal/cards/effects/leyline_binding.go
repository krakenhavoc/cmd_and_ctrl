package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline Binding — Enchantment {5}{W} (Marvel Universe, #4460):
//
//	"Flash
//	 Domain — This spell costs {1} less to cast for each basic land type
//	 among lands you control.
//	 When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Banishing Light's exile pair (exileUntilThisLeavesTriggers, CR 610.3)
// under flash. The domain count (an ability word, CR 207.2c: no rules
// meaning of its own) is the number of the five basic land types (CR 205.3i) found on
// the lands its caster controls, each counted once however many lands have
// it, read from the layered land (a Dryad of the Ilysian Grove adds all
// five; a Blood Moon makes every nonbasic a Mountain). The reduction is
// generic mana only and never goes below the printed colored cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b49642f6-20bd-468d-9dac-e578f9ac3091",
		Name:            "Leyline Binding",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		SelfCostModifiers: []game.CostModifier{
			DomainCostReduction("Domain — This spell costs {1} less to cast for each basic land type among lands you control"),
		},
		Triggered: exileUntilThisLeavesTriggers("Leyline Binding"),
	})
}
