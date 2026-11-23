package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brightfield Mustang — Creature — Horse Mount {3}{W}:
//
//	"Whenever this creature attacks while saddled, untap it and put a
//	 +1/+1 counter on it.
//	 Saddle 1"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b05e355-38a5-4307-9d8b-e99ec30f8ed3",
		Name:         "Brightfield Mustang",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(1)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Brightfield Mustang — untap it and put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
					return err
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		},
	})
}
