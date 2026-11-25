package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scarecrone — Artifact Creature — Scarecrow {3}, 1/2 (Jumpstart, #5510):
//
//	"{1}, Sacrifice a Scarecrow: Draw a card.
//	 {4}, {T}: Return target artifact creature card from your graveyard to the
//	 battlefield."
//
// The sacrifice may be Scarecrone itself, as printed. The return is
// Gravewaker's untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f45716dc-c613-4b14-865d-419395fb02f6",
		Name:         "Scarecrone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a Scarecrow: Draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "a Scarecrow", effects.Subtype("Scarecrow"))),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}, {
			Label:   "{4}, {T}: Return target artifact creature card from your graveyard to the battlefield.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target artifact creature card from your graveyard", effects.Artifact(), effects.Creature(), effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
				return nil
			},
		}},
	})
}
