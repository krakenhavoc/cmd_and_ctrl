package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aquastrand Spider — Creature — Spider Mutant {1}{G}, 0/0 (Dissension, #5598):
//
//	"Graft 2 (This creature enters with two +1/+1 counters on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)
//	 {G}: Target creature with a +1/+1 counter on it gains reach until
//	 end of turn. (It can block creatures with flying.)"
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
// The grant is disGraftGrantRow, Helium Squirter's flying row with
// reach. The target must have a +1/+1 counter as it is chosen and again
// as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	replacements, triggered := effects.DisGraft("Aquastrand Spider", 2)
	effects.Register(effects.Spec{
		OracleID:     "cea79fa0-2a76-41f3-846f-4d1a6ff5f629",
		Name:         "Aquastrand Spider",
		Completeness: effects.CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Activated:    []effects.ActivatedAbility{effects.DisGraftGrantRow("{G}", "reach", "Aquastrand Spider")},
	})
}
