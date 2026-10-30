package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flash Thompson, Spider-Fan — Legendary Creature — Human Citizen {1}{W},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 When Flash Thompson enters, choose one or both —
//	 • Heckle — Tap target creature.
//	 • Hero Worship — Untap target creature."
//
// A modal enters trigger (Giant-Sized Flying Ant's shape) with "one or
// both" (ChooseOneOrMore). Each bullet has its own target (CR 115.3), so
// both may name the same creature and run in printed order.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Flash Thompson — choose one or both", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOneOrMore(
		ModeDoing("Heckle — Tap target creature.", TargetCreature("target creature"), TapTheModesTarget),
		ModeDoing("Hero Worship — Untap target creature.", TargetCreature("target creature"), UntapTheModesTarget),
	)
	Register(Spec{
		OracleID:        "f536a518-160e-4267-89da-3d34accc6d2c",
		Name:            "Flash Thompson, Spider-Fan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
