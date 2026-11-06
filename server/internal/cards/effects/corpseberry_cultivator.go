package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corpseberry Cultivator — Creature — Squirrel Warlock {1}{B/G}{B/G},
// 2/3 (Bloomburrow, #4467):
//
//	"At the beginning of combat on your turn, you may forage. (Exile
//	 three cards from your graveyard or sacrifice a Food.)
//	 Whenever you forage, put a +1/+1 counter on this creature."
//
// The combat trigger is MayForage with nothing after it (#5184): the
// controller is asked only when they can forage. "Whenever you forage"
// is WheneverYouForage, off game.EventForage, so it sees every forage
// its controller makes — this one, another card's "Forage" cost, or a
// spell's forage branch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ded56df-c7d5-4043-85df-410f47219faa",
		Name:         "Corpseberry Cultivator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Corpseberry Cultivator — you may forage",
				func(g *game.Game, item *game.StackItem) error {
					return MayForage(NewContext(g, item), "Corpseberry Cultivator — forage?", mayForageNothingMore)
				}),
			WheneverYouForage("Corpseberry Cultivator — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
