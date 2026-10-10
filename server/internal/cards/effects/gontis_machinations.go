package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gonti's Machinations — {B} Enchantment:
//
//	"Whenever you lose life for the first time each turn, you get {E}.
//	 (You get an energy counter. Damage causes loss of life.)
//	 Pay {E}{E}, Sacrifice this enchantment: Each opponent loses 3 life.
//	 You gain life equal to the life lost this way."
//
// The trigger is WheneverYouLoseLifeForTheFirstTimeEachTurn (#2540): the
// turn tally says whether a loss is the turn's first, so several
// creatures hitting you in one combat damage step give one energy.
//
// The drain is Exsanguinate's body at 3
// (b21DrainEachOpponentAndGainTheTotal): "life lost this way" is what
// the opponents actually lost, after any replacement. The energy and
// the sacrifice are costs (CR 118.3, 601.2h), paid before anything
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e6238d8-bd6a-4952-a7dd-9c41a78a2cba",
		Name:         "Gonti's Machinations",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				WheneverYouLoseLifeForTheFirstTimeEachTurn("Gonti's Machinations — you get "+EnergySymbols(1),
					Do(GetEnergy{N: 1})),
				game.Purpose{Energy: 1}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay {E}{E}, Sacrifice this enchantment: Each opponent loses 3 life. You gain life equal to the life lost this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(PayEnergy(2), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b21DrainEachOpponentAndGainTheTotal(g, item, 3)
			},
		}},
	})
}
