package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Own the Day — Sorcery {4}{R}{R} (Star Trek Commander, #4974):
//
//	"Battlefront — This spell costs {1} less to cast for each opponent
//	 you attacked this turn.
//	 Untap all creatures that attacked this turn. After this main phase,
//	 there is an additional combat phase followed by an additional main
//	 phase."
//
// Battlefront is an ability word. The discount is a self cost modifier
// (CR 601.2f, Spec.SelfCostModifiers) counting the opponents you attacked
// this turn as you cast it (trekB37OpponentsAttackedThisTurn): each
// opponent once, a planeswalker or battle no one. It reduces generic mana
// only, so the floor is {R}{R}. The resolution is Relentless Assault's,
// the same function (untapAttackersThenCombatAndMain): the extra phases
// go on the current turn, and only when it resolves in a main phase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a9bfeb2-b4b3-4c69-9739-a878c2704f46",
		Name:         "Own the Day",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				return trekB37OpponentsAttackedThisTurn(q.Game, q.Controller)
			}, "Battlefront — This spell costs {1} less to cast for each opponent you attacked this turn."),
		},
		OnResolve: untapAttackersThenCombatAndMain,
	})
}
