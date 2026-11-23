package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plaxcaster Frogling — Creature — Frog Mutant {1}{G}{U}, 0/0 (Dissension, #5598):
//
//	"Graft 3 (This creature enters with three +1/+1 counters on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)
//	 {2}: Target creature with a +1/+1 counter on it gains shroud until
//	 end of turn. (It can't be the target of spells or abilities.)"
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
// The grant is disGraftGrantRow with shroud.
//
// No simplification.
func init() {
	replacements, triggered := effects.DisGraft("Plaxcaster Frogling", 3)
	effects.Register(effects.Spec{
		OracleID:     "8c1ea1a5-5966-4d14-8361-ffc6a8d1bd79",
		Name:         "Plaxcaster Frogling",
		Completeness: effects.CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Activated:    []effects.ActivatedAbility{effects.DisGraftGrantRow("{2}", "shroud", "Plaxcaster Frogling")},
	})
}
