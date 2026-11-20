package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "57170e91-6ebe-4b14-bf56-96d8cf6a4d03",
		Name:            "Licia, Sanguine Tribune",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return b15LifeGainedThisTurn(q.Game, q.Controller)
			}, "This spell costs {1} less to cast for each 1 life you gained this turn."),
		},
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Condition: AllConditions(DuringYourTurn(), OncePerTurnActivation(label)),
			Cost:      game.AbilityCost{Life: 5},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 3}.Apply(NewContext(g, item))
			},
		}},
	})
}
