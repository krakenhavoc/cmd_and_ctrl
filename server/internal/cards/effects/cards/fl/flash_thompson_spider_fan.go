package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	enters := effects.WhenThisEnters("Flash Thompson — choose one or both", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOneOrMore(
		effects.ModeDoing("Heckle — Tap target creature.", effects.TargetCreature("target creature"), effects.TapTheModesTarget),
		effects.ModeDoing("Hero Worship — Untap target creature.", effects.TargetCreature("target creature"), effects.UntapTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:        "f536a518-160e-4267-89da-3d34accc6d2c",
		Name:            "Flash Thompson, Spider-Fan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
