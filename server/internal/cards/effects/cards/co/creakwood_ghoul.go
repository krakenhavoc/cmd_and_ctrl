package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Creakwood Ghoul — {4}{B} Creature — Plant Zombie, 3/3:
//
//	"{B/G}{B/G}: Exile target card from a graveyard. You gain 1 life."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94c47804-bb64-4a91-85c5-9241c39685aa",
		Name:         "Creakwood Ghoul",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B/G}{B/G}: Exile target card from a graveyard. You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.ManaCost("{B/G}{B/G}"),
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
				return effects.GainLife{Amount: 1}.Apply(ctx)
			},
		}},
	})
}
