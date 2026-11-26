package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thoughtcutter Agent — Artifact Creature — Human Rogue {U}{B}, 1/1 (Shards
// of Alara, S82 OLD batch 512, #5588):
//
//	"{U}{B}, {T}: Target player loses 1 life and reveals their hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06ada754-da78-4eda-8be7-3ac4f06115c4",
		Name:         "Thoughtcutter Agent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}, {T}: Target player loses 1 life and reveals their hand.",
			Cost:    effects.Plus(effects.ManaCost("{U}{B}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				p := effects.TargetedPlayerOf(ctx)
				if p == uuid.Nil {
					return nil
				}
				if err := (effects.GainLife{Player: p, Amount: -1}).Apply(ctx); err != nil {
					return err
				}
				g.RevealHandForEffect(p)
				return nil
			},
		}},
	})
}
