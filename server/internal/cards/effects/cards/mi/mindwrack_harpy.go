package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindwrack Harpy — Enchantment Creature — Harpy {3}{B}, 3/2 (Commander
// Masters):
//
//	"Flying
//	 At the beginning of combat on your turn, each player mills three cards."
//
// The trigger is the beginning-of-combat step trigger of your turn; each
// player, in turn order starting with the active player (CR 101.4), mills
// three (a short library mills what it has and does not lose, CR 701.17).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f20f27d1-2c50-41d5-b1af-44a0e7ca2e00",
		Name:            "Mindwrack Harpy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Mindwrack Harpy — each player mills three cards", mindwrackHarpyMill),
		},
	})
}

func mindwrackHarpyMill(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, p := range effects.GapB34TableInTurnOrder(ctx) {
		if err := (effects.MillCards{Player: p, N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
