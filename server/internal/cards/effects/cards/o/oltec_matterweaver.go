package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oltec Matterweaver — Creature — Human Artificer {2}{W}, 3/2 (The Big
// Score, #4468):
//
//	"Whenever you cast a creature spell, choose one —
//	 • Create a 1/1 colorless Gnome artifact creature token.
//	 • Create a token that's a copy of target artifact token you control."
//
// A modal cast trigger; the copy mode is only choosable with an artifact
// token to target (CR 700.2b) and copies what the token was created as
// (CR 707.2).
//
// No simplification.
func init() {
	t := effects.WheneverYouCast(effects.Creature(), "Oltec Matterweaver — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeSteps("Create a 1/1 colorless Gnome artifact creature token.", nil,
			effects.CreateToken{Template: effects.TokenCard("1/1 colorless Gnome artifact"), N: 1}),
		effects.ModeDoing("Create a token that's a copy of target artifact token you control.", effects.B35ArtifactTokenYouControl(), effects.B35CopyTargetTokenModeBody),
	)
	effects.Register(effects.Spec{
		OracleID:     "3cb91696-242f-4e91-806b-61b1bb45a20c",
		Name:         "Oltec Matterweaver",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
