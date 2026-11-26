package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Font of Return — Enchantment {1}{B} (Journey into Nyx, S82 OLD, #5558):
//
//	"{3}{B}, Sacrifice this enchantment: Return up to three target creature
//	 cards from your graveyard to your hand."
//
// Soul of Innistrad's return of up to three creature cards, with the
// sacrifice as a cost. The targets are chosen as the ability goes on the
// stack, so the Font is already gone when they are returned, and whichever are
// still legal return (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8805aa6-fa5f-4c70-ad90-8502611bef5a",
		Name:         "Font of Return",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, Sacrifice this enchantment: Return up to three target creature cards from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("up to three target creature cards from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 3),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
			},
		}},
	})
}
