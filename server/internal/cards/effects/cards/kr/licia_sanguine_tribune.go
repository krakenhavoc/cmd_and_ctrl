package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Licia, Sanguine Tribune — Legendary Creature — Vampire Soldier
// {5}{R}{W}{B}, 4/4 (Commander 2017, #5533):
//
//	"This spell costs {1} less to cast for each 1 life you gained this
//	 turn.
//	 First strike, lifelink
//	 Pay 5 life: Put three +1/+1 counters on Licia. Activate only during
//	 your turn and only once each turn."
//
// The reduction is a self cost modifier (Sanguine Indulgence's read of the
// turn tally's life gained by the caster) sized per cast, so it takes only
// generic mana off (CR 601.2f). The activated ability pays 5 life as its
// cost (CR 119.4) and is gated by Snowslope Hunter's pair of timing limits,
// counted per announcement (CR 602.1b). A Licia that left in response is a
// new object and takes no counters (CR 400.7).
//
// No simplification.
func init() {
	const label = "Pay 5 life: Put three +1/+1 counters on Licia. Activate only during your turn and only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "57170e91-6ebe-4b14-bf56-96d8cf6a4d03",
		Name:            "Licia, Sanguine Tribune",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return effects.B15LifeGainedThisTurn(q.Game, q.Controller)
			}, "This spell costs {1} less to cast for each 1 life you gained this turn."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Condition: effects.AllConditions(effects.DuringYourTurn(), effects.OncePerTurnActivation(label)),
			Cost:      game.AbilityCost{Life: 5},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
