package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wandermare — Creature — Horse {1}{G}{W}, 3/3 (Throne of Eldraine, #5516):
//
//	"Whenever you cast a creature spell that has an Adventure, put a +1/+1
//	 counter on this creature. (It doesn't need to have gone on the
//	 Adventure first.)"
//
// A creature spell is the card's permanent half, so the Adventure half never
// counts; the card has an Adventure whether cast from hand or from exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c25470c-7be0-495c-b756-6a4fc230d660",
		Name:         "Wandermare",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(s82OldB144CreatureWithAdventure),
				"Wandermare — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

func s82OldB144CreatureWithAdventure(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && effects.CardHasAdventure(c)
}
