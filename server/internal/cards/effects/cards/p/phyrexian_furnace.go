package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Furnace — Artifact {1} (Weatherlight):
//
//	"{T}: Exile the bottom card of target player's graveyard.
//	 {1}, Sacrifice this artifact: Exile target card from a graveyard. Draw a
//	 card."
//
// The bottom card is the one that has been in the pile longest (CR 404.2,
// #6194), read as the ability resolves; a target player with an empty graveyard
// loses nothing. The second ability's draw is part of the same resolution, so it
// is skipped with the exile when the target card is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4aafb7e7-b9fc-44ae-9c54-da11fd9dcc17",
		Name:         "Phyrexian Furnace",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Exile the bottom card of target player's graveyard.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard},
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						if id := g.BottomGraveyardCardForEffect(t.ID); id != uuid.Nil {
							return effects.ExileTarget{Target: id}.Apply(ctx)
						}
					}
					return nil
				},
			},
			{
				Label:   "{1}, Sacrifice this artifact: Exile target card from a graveyard. Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.DrawCards{N: 1}.Apply(ctx)
				},
			},
		},
	})
}
