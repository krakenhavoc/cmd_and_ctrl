package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squad Captain — {4}{W} Creature — Human Soldier 2/2 (Foundations
// Jumpstart):
//
//	"Vigilance
//	 This creature enters with a +1/+1 counter on it for each other
//	 creature you control."
//
// A CR 614.1c self-replacement counted as the creature enters
// (b19EntersWithCountersCounted), so a creature that died in response is
// not counted.
//
// The count is Crowd Control Warden's shared "other creatures you control".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "865812c2-1838-4015-bf0d-990af76f7000",
		Name:            "Squad Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, crowdControlWardenX, "Squad Captain enters with a +1/+1 counter on it for each other creature you control"),
		},
	})
}
