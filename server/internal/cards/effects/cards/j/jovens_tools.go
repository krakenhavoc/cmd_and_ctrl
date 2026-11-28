package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Joven's Tools — Artifact {6}, (Homelands, S82 OLD b821, #5635):
//
//	"{4}, {T}: Target creature can't be blocked this turn except by Walls."
//
// Rime Transfusion's turn-scoped "except by" rule (CantBeBlockedThisTurnExceptBy)
// pinned to the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1935140-b038-4a81-98f7-28d2c137124a",
		Name:         "Joven's Tools",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Target creature can't be blocked this turn except by Walls.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.CantBeBlockedThisTurnExceptBy{
					Target:   id,
					Subtypes: []string{"Wall"},
					Text:     "Walls",
					Label:    "Joven's Tools — can't be blocked except by Walls",
				}.Apply(ctx)
			},
		}},
	})
}
