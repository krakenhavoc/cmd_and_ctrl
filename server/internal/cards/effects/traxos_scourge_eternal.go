package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traxos, Scourge Eternal — Legendary Artifact Creature — Dragon
// Construct {4}, 5/4:
//
//	"Trample
//	 Traxos doesn't untap during your untap step.
//	 Whenever you cast an artifact or creature spell, untap Traxos."
//
// The untap-step restriction is CR 502.3 (Island Fish Jasconius' shape);
// the untap is a trigger on casting, so it uses the stack and an
// opponent can respond. A Traxos that left before the trigger resolves
// is not untapped (b31UntapSelf checks).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "acf4abde-4034-4fcf-b531-78c935aecd8c",
		Name:                  "Traxos, Scourge Eternal",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{"trample"},
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(rfCreatureFArtifactOrCreature(),
				"Traxos, Scourge Eternal — untap Traxos", b31UntapSelf),
		},
	})
}
