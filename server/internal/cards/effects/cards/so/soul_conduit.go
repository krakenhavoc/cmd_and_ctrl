package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Conduit — Artifact {6}:
//
//	"{6}, {T}: Two target players exchange life totals."
//
// With either target gone by resolution nothing is exchanged: a player can't
// exchange life totals with no one (CR 701.12a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b48befc5-a9f2-4a64-a16d-c883b034ea9c",
		Name:         "Soul Conduit",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}: Two target players exchange life totals.",
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost()),
			Targets: effects.TargetPlayer("two target players").WithCount(2, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var players []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						players = append(players, t.ID)
					}
				}
				if len(players) != 2 {
					return nil
				}
				return effects.ExchangeLifeTotals(ctx, players[0], players[1])
			},
		}},
	})
}
