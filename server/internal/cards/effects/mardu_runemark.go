package effects

// Mardu Runemark — Enchantment — Aura {2}{R} (Fate Reforged,
// #5552):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Enchanted creature has first strike as long as you control a white
//	 or black permanent."
//
// Enchant creature is the Aura's target (EnchantCreature). The Runemark
// statics (frfRunemark): +2/+2 in layer 7c always, and first strike in layer
// 6 while you control a white or black permanent, re-read on every
// recompute. "You" is the Aura's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a912834b-e4bc-4ccc-ae13-915ebe5acd00",
		Name:         "Mardu Runemark",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       frfRunemark("first strike", "W", "B"),
	})
}
