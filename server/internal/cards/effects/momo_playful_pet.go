package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	momo := WhenThisLeaves("Momo, Playful Pet — choose one", func(*game.Game, *game.StackItem) error { return nil })
	momo.Modes = ChooseOne(
		ModeSteps("Create a Food token.", nil, CreateToken{Template: FoodToken(), N: 1}),
		ModeDoing("Put a +1/+1 counter on target creature you control.",
			TargetCreature("target creature you control", YouControl()),
			PlusOneCountersOnTheModesTarget(1)),
		ModeWithPurpose(ModeSteps("Scry 2.", nil, Scry{N: 2}), game.Purpose{Scry: 2}),
	)
	Register(Spec{
		OracleID:        "85ee8ab6-8db5-45d7-85ae-64838771146a",
		Name:            "Momo, Playful Pet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered:       []game.TriggeredAbility{momo},
	})
}
