package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fast Forward — {4}{R} Sorcery (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"This spell costs {1} less to cast for each opponent you attacked
//	 this turn.
//	 Goad all creatures your opponents control. (Until your next turn,
//	 those creatures attack each combat if able and attack a player
//	 other than you if able.)"
//
// The reduction counts the opponents a creature of yours was declared
// attacking this turn (CR 601.2f; it touches the generic part only, so
// three opponents make it {1}{R}). An attack on a planeswalker or battle
// is not an attack on that player. The goad is Disrupt Decorum's, on
// every creature an opponent controls as it resolves (CR 701.15).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c92f78c4-c8ec-4bdb-85f4-3a206b709f0f",
		Name:         "Fast Forward",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return OpponentsYouAttackedThisTurn(q.Game, q.Controller)
			}, "This spell costs {1} less to cast for each opponent you attacked this turn."),
		},
		OnResolve: goadAllCreaturesYourOpponentsControl,
	})
}
