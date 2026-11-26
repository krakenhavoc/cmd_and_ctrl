package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toolcraft Exemplar — Creature — Dwarf Artificer {W}, 1/1 (Kaladesh, #5539):
//
//	"At the beginning of combat on your turn, if you control an artifact,
//	 this creature gets +2/+1 until end of turn. If you control three or more
//	 artifacts, it also gains first strike until end of turn."
//
// "If you control an artifact" is an intervening if (CR 603.4): read as the
// step begins and again as the ability resolves, through the layered view.
// The first strike is part of the same ability, so it needs only the three
// artifacts as it resolves (the first condition having held to get there),
// and both effects end at cleanup (CR 514.2).
//
// No simplification.
func init() {
	label := "Toolcraft Exemplar — this creature gets +2/+1 until end of turn; with three or more artifacts it also gains first strike"
	trigger := effects.InterveningIfYou("you control an artifact",
		func(g *game.Game, you uuid.UUID) bool { return effects.S82OldB260ArtifactsYouControl(g, you) >= 1 },
		effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
			if err := effects.ThisGetsUntilEndOfTurn(2, 1, "Toolcraft Exemplar — +2/+1")(g, item); err != nil {
				return err
			}
			if effects.S82OldB260ArtifactsYouControl(g, item.Controller) < 3 {
				return nil
			}
			return effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Toolcraft Exemplar — first strike")(g, item)
		}))
	effects.Register(effects.Spec{
		OracleID:     "c51b0667-a24e-44c0-bc95-bf61b393f3cf",
		Name:         "Toolcraft Exemplar",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
