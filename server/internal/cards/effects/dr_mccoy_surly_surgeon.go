package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dr. McCoy, Surly Surgeon — {2}{W}{W} Legendary Creature — Human Doctor
// 2/2 (Star Trek, #4973):
//
//	"When Dr. McCoy enters, put a +1/+1 counter on another target
//	 creature you control.
//	 If another creature you control with a +1/+1 counter on it would be
//	 destroyed, instead remove all +1/+1 counters from it and all damage
//	 already dealt to it is healed."
//
// The enters trigger is Foggy Nelson's target clause with the counter
// alone. The second ability is a CR 614.1a replacement of a destruction
// (trekB2CountersInsteadOfDestruction): an effect's destroy and the
// state-based destruction for lethal or deathtouch damage alike, but not a
// sacrifice or zero toughness. The creature stays, loses every +1/+1
// counter, and its marked damage is healed (CR 701.69a). Once it has no
// +1/+1 counter, a second destruction is not replaced. With a regeneration
// shield or umbra armor also in reach, the creature's controller picks
// which applies (CR 616.1).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Dr. McCoy, Surly Surgeon — put a +1/+1 counter on another target creature you control", PlusOneCounterOnTheTarget)
	enters.Targets = Another(TargetCreature("another target creature you control", YouControl()))
	Register(Spec{
		OracleID:     "df8a0737-5841-4c17-b567-e9296541c3c9",
		Name:         "Dr. McCoy, Surly Surgeon",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Replacements: []game.ReplacementEffect{
			trekB2CountersInsteadOfDestruction("Dr. McCoy, Surly Surgeon: remove its +1/+1 counters and heal it instead"),
		},
	})
}
