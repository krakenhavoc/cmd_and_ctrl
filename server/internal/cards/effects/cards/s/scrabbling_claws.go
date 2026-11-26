package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrabbling Claws — Artifact {1} (Ravnica Allegiance):
//
//	"{T}: Target player exiles a card from their graveyard.
//	 {1}, Sacrifice this artifact: Exile target card from a graveyard. Draw a card."
//
// Relic of Progenitus first ability: the target player chooses the card. The
// second ability targets the card (CR 602.2b); if it is gone by resolution
// nothing happens, the draw included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4538ec2c-867e-412b-b6e3-a57f8e29ba84",
		Name:         "Scrabbling Claws",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target player exiles a card from their graveyard.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard},
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.OldB178TargetExilesOwnGraveyardCard,
			},
			{
				Label:   "{1}, Sacrifice this artifact: Exile target card from a graveyard. Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				},
			},
		},
	})
}
