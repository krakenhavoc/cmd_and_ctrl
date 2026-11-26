package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hired Torturer — Creature — Human Rogue {2}{B}, 2/3 (Dragon's Maze, S82 OLD
// b406, #5564):
//
//	"Defender
//	 {3}{B}, {T}: Target opponent loses 2 life, then reveals a card at random
//	 from their hand."
//
// The life loss comes first and the reveal is a random pick from the target's
// hand (Friendly Fire's keyed pick), announced to the table. Both skip a
// target that is no longer legal as the ability resolves (CR 608.2b); an empty
// hand reveals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1ad2fd3-9f39-49e7-92c5-57ab0e1da5b7",
		Name:            "Hired Torturer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, {T}: Target opponent loses 2 life, then reveals a card at random from their hand.",
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim := effects.TargetedPlayer(ctx)
				if victim == uuid.Nil {
					return nil
				}
				if err := g.ChangePlayerLifeForEffect(ctx.Source(), victim, -2); err != nil {
					return err
				}
				effects.S82OldB406RevealRandomCardFromHand(ctx, victim, "Hired Torturer")
				return nil
			},
		}},
	})
}
