package effects

// Temur Runemark — Enchantment — Aura {2}{G} (Fate Reforged,
// #5552):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Enchanted creature has trample as long as you control a blue or red
//	 permanent."
//
// Enchant creature is the Aura's target (EnchantCreature). The Runemark
// statics (frfRunemark): +2/+2 in layer 7c always, and trample in layer 6
// while you control a blue or red permanent, re-read on every recompute.
// "You" is the Aura's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "590c2455-90f7-4ff6-a6cf-5d0d25575cda",
		Name:         "Temur Runemark",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       frfRunemark("trample", "U", "R"),
	})
}
