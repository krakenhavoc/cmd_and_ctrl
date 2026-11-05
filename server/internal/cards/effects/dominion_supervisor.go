package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dominion Supervisor — Creature — Vorta Doctor {2}{G}, 3/1 (Star Trek,
// #4973):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 Creatures you control with counters on them have trample."
//
// The enters trigger targets any creature, the Supervisor itself
// included, and does nothing if the target is gone (CR 608.2b). The
// static is Nev, the Practical Dean's layer-6 grant read off the
// permanent's counters, any kind.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d151c8ed-58bf-4b82-a993-1d81f789aa0e",
		Name:         "Dominion Supervisor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Dominion Supervisor — put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
		},
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(socB9CreatureYouControlWithACounter, "trample"), "Creatures you control with counters on them have trample."),
		},
	})
}
