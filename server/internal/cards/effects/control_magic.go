package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Control Magic — Enchantment — Aura for {2}{U}{U}:
//
//	"Enchant creature
//	 You control enchanted creature."
//
// Mind Control a mana cheaper and thirty years older, and identical
// to it in every line of code. It is in this batch for what the
// SECOND control-changer proves rather than for the card itself: two
// layer-2 effects on one creature sort by CR 613.7 timestamp and the
// later one wins, and when the later one leaves, the EARLIER one
// takes over rather than control going home to the original
// controller. Nothing in either card file says any of that — it falls
// out of the recompute's sort — and until there were two of these,
// nothing could test it.
//
// Everything Mind Control's file records applies here verbatim:
// control reverts by itself when the Aura leaves, CR 302.6 gives the
// creature summoning sickness under its new controller, CR 506.4
// removes it from combat, and CR 704.5n puts the Aura in the
// graveyard if the creature stops being one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd0d7141-46d2-4aa3-bc77-6b3b4513803e",
		Name:         "Control Magic",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			ControlAttachedBySource(),
		},
	})
}
