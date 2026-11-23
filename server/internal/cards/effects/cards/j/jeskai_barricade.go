package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeskai Barricade — Creature — Wall {1}{W}, 0/4 (Fate Reforged,
// #5552):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Defender
//	 When this creature enters, you may return another target creature
//	 you control to its owner's hand."
//
// Flash and defender are printed. The entry trigger is Auramancer's optional
// targeted shape: "another" excludes the Barricade by object, the target is
// re-checked as it resolves, and with no other creature you control the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a2fa0a6-1324-4714-8896-6bb97640578a",
		Name:            "Jeskai Barricade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "defender"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Jeskai Barricade — return another target creature you control to its owner's hand", effects.BounceChosenTarget),
					"Jeskai Barricade — return another creature you control to its owner's hand?"),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
