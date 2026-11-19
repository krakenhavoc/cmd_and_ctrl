package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relief Captain — Creature — Kor Knight Ally {2}{W}{W}, 3/2 (Foundations
// Jumpstart, S82 BNS):
//
//	"When this creature enters, support 3. (Put a +1/+1 counter on each of
//	 up to three other target creatures.)"
//
// Support is the keyword action (support.go, CR 701.41a): up to three other
// target creatures, whoever controls them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97e088a3-65eb-42f5-85b8-6aeb770a3f56",
		Name:         "Relief Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersSupport("Relief Captain", 3, true),
		},
	})
}
