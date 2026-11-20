package effects

// Jeskai Runemark — Enchantment — Aura {2}{U} (Fate Reforged,
// #5552):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Enchanted creature has flying as long as you control a red or white
//	 permanent."
//
// Enchant creature is the Aura's target (EnchantCreature). The Runemark
// statics (frfRunemark): +2/+2 in layer 7c always, and flying in layer 6
// while you control a red or white permanent, re-read on every recompute.
// "You" is the Aura's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5fe13418-c63c-4f3c-9a42-7ab06949e3cd",
		Name:         "Jeskai Runemark",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       frfRunemark("flying", "R", "W"),
	})
}
