package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bhaal, Lord of Murder — {2}{B}{R}{G} Legendary Creature — God 4/4
//
//	"As long as your life total is less than or equal to half your starting
//	 life total, Bhaal has indestructible.
//	 Whenever another nontoken creature you control dies, put a +1/+1 counter
//	 on target creature and goad it."
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b); a creature that left in response gets neither the
// counter nor the goad. The goad lasts until your next turn (CR 701.15a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d2dd78c-103c-4286-a3a1-e471dc77ed91",
		Name:         "Bhaal, Lord of Murder",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GapB214SelfIndestructibleAtHalfLife("Bhaal")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventLTB, effects.HofriGhostforgeAnotherNontokenCreatureDied,
					"Bhaal, Lord of Murder — put a +1/+1 counter on target creature and goad it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
						return effects.GoadAndScheduleClear(ctx, []uuid.UUID{id})
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
