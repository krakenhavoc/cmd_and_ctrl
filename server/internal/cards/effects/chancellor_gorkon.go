package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chancellor Gorkon — Legendary Creature — Klingon Noble {3}{W}{B}, 5/3
// (Star Trek Commander, #4974):
//
//	"Vigilance
//	 When Chancellor Gorkon dies, exile up to two target creatures."
//
// Vigilance is printed and enforced. The dies trigger (CR 700.4) asks for
// zero, one or two target creatures as it is put on the stack, anyone's;
// Gorkon is in the graveyard by then, so he cannot be one of them. Each
// target still legal at resolution is exiled; one that became illegal is
// left alone (CR 608.2b), and with none left the trigger does nothing.
//
// No simplification.
func init() {
	dies := WhenThisDies("Chancellor Gorkon — exile up to two target creatures", exileEveryLegalCardTarget)
	dies.Targets = TargetCreature("up to two target creatures").WithCount(0, 2)
	Register(Spec{
		OracleID:        "24d194d3-f83a-496b-921a-77c09c68b0e1",
		Name:            "Chancellor Gorkon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{dies},
	})
}
