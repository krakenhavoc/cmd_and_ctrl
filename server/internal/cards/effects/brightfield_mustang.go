package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brightfield Mustang — Creature — Horse Mount {3}{W}:
//
//	"Whenever this creature attacks while saddled, untap it and put a
//	 +1/+1 counter on it.
//	 Saddle 1"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b05e355-38a5-4307-9d8b-e99ec30f8ed3",
		Name:         "Brightfield Mustang",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(1)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Brightfield Mustang — untap it and put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		},
	})
}
