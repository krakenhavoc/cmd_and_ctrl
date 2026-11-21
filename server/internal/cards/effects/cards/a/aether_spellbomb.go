package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Spellbomb — Artifact {1} (EDHREC rank 2761):
//
//	"{U}, Sacrifice this artifact: Return target creature to its
//	 owner's hand.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The Mirrodin spellbomb: a one-mana artifact that is either an
// Unsummon or a cantrip, and an artifact-count for the deck that
// wants one either way. Two activated abilities, each sacrificing
// the Spellbomb as its cost, so a dies-trigger on it (Chromatic Star
// style) lands above the ability. The bounce is targeted; a creature
// that left in response makes the ability do nothing, and the
// Spellbomb is gone either way, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b033a0a-c1ae-44d7-9662-72cbbfda024b",
		Name:         "Aether Spellbomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, Sacrifice this artifact: Return target creature to its owner's hand.",
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BounceToHand{Target: id}.Apply(ctx)
				},
			},
			{
				Label:   "{1}, Sacrifice this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
