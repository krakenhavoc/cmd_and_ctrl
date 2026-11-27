package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Village Cannibals — Creature — Human {2}{B}, 2/2 (Innistrad, S82 older sets
// batch 449, #5572):
//
//	"Whenever another Human creature dies, put a +1/+1 counter on this
//	 creature."
//
// Any controller's Human, the Cannibals excluded, read as it last existed
// (CR 603.10a), so a creature that was a Human only through an effect still
// counts. A changeling is every type and counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "973504d6-e2e4-49c4-86cb-69f31d66388b",
		Name:         "Village Cannibals",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.S82OldB449AnotherSubtypeCreatureDied("Human"),
				"Village Cannibals — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
