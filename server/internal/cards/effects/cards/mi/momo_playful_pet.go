package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Momo, Playful Pet — Legendary Creature — Lemur Bat Ally {W}, 1/1:
//
//	"Flying, vigilance
//	 When Momo leaves the battlefield, choose one —
//	 • Create a Food token.
//	 • Put a +1/+1 counter on target creature you control.
//	 • Scry 2."
//
// Any exit triggers it, not only dying. No simplification.
func init() {
	momo := effects.WhenThisLeaves("Momo, Playful Pet — choose one", func(*game.Game, *game.StackItem) error { return nil })
	momo.Modes = effects.ChooseOne(
		effects.ModeSteps("Create a Food token.", nil, effects.CreateToken{Template: effects.FoodToken(), N: 1}),
		effects.ModeDoing("Put a +1/+1 counter on target creature you control.",
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.PlusOneCountersOnTheModesTarget(1)),
		effects.ModeWithPurpose(effects.ModeSteps("Scry 2.", nil, effects.Scry{N: 2}), game.Purpose{Scry: 2}),
	)
	effects.Register(effects.Spec{
		OracleID:        "85ee8ab6-8db5-45d7-85ae-64838771146a",
		Name:            "Momo, Playful Pet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered:       []game.TriggeredAbility{momo},
	})
}
