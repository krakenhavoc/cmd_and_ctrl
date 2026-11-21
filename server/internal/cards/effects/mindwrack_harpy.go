package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f20f27d1-2c50-41d5-b1af-44a0e7ca2e00",
		Name:            "Mindwrack Harpy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Mindwrack Harpy — each player mills three cards", mindwrackHarpyMill),
		},
	})
}

func mindwrackHarpyMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, p := range gapB34TableInTurnOrder(ctx) {
		if err := (MillCards{Player: p, N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
