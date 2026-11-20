package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vigean Hydropon — Creature — Plant Mutant {1}{G}{U}, 0/0 (Dissension, #5598):
//
//	"Graft 5 (This creature enters with five +1/+1 counters on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)
//	 This creature can't attack or block."
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
// The restriction is Grakk the Pacifist's RestrictSelf.
//
// No simplification.
func init() {
	replacements, triggered := disGraft("Vigean Hydropon", 5)
	Register(Spec{
		OracleID:     "3eab4c00-a8ae-4912-bee1-68376de700a3",
		Name:         "Vigean Hydropon",
		Completeness: CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Static:       []game.StaticAbility{RestrictSelf(game.CantAttackOrBlock)},
	})
}
