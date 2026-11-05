package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcade Cabinet — {3} Artifact (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"When this artifact enters, put a +1/+1 counter on each of up to four
//	 target creatures.
//	 {2}, {T}, Sacrifice a token: Double the number of each kind of
//	 counter on target creature."
//
// The enters trigger is Agent Bishop's "each of up to N target creatures"
// (putPlusOneCounterOnEachLegalTarget), four this time: any creatures,
// anyone's, and a target that became illegal is skipped (CR 608.2b). The
// activation is Vorel of the Hull Clade's doubler
// (b30DoubleCountersOnFirstLegalTarget: each kind read once, then put on
// again through the ordinary placement, so a counter replacement applies,
// as the printed ruling says) on one target creature, with Fountainport's
// "Sacrifice a token" as a cost (CR 602.2b). An artifact, so the tap needs
// no wait (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8e55287-f0df-46f2-a41d-27bea90216cc",
		Name:         "Arcade Cabinet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Arcade Cabinet — put a +1/+1 counter on each of up to four target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to four target creatures").WithCount(0, 4)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a token: Double the number of each kind of counter on target creature.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), b08SacrificeAToken()),
			Targets: TargetCreature("target creature"),
			Effect:  b30DoubleCountersOnFirstLegalTarget,
		}},
	})
}
