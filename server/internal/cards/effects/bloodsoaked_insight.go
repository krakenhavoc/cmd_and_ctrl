package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodsoaked Insight // Sanguine Morass — Modal DFC. Front: Sorcery
// {5}{B/R}{B/R}:
//
//	"This spell costs {1} less to cast for each 1 life your opponents
//	 have lost this turn.
//	 Target opponent exiles the top three cards of their library. Until
//	 the end of your next turn, you may play those cards. If you cast a
//	 spell this way, mana of any type can be spent to cast it."
//
// Back: Land (mdfc_lands.go).
//
// The discount is a self cost modifier read off the turn tally of life
// lost (the same count Rakdos, Lord of Riots reads). The exile is the
// impulse shape on an opponent's library: the target exiles their own
// top three, the caster is granted the play permission, the window is
// "until the end of your next turn" (a seat-turn duration, ADR 0063), and
// AnyType is "mana of any type can be spent" (CR 609.4b). It covers a
// spell cast from exile and not a land, which has no cost. No
// simplification.
func init() {
	Register(Spec{
		OracleID:     "c52fc8a1-43c6-41f8-b010-03be7c89ef1d",
		Name:         "Bloodsoaked Insight",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(
				func(q game.CostQuery) int { return opponentLifeLostThisTurn(q.Game, q.Controller) },
				"This spell costs {1} less to cast for each 1 life your opponents have lost this turn."),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			opp := TargetedPlayer(ctx)
			if p := ctx.PlayerByID(opp); p == nil || p.Eliminated {
				return nil
			}
			window := ctx.Game.UntilEndOfYourNextTurnDuration(item.Controller)
			return ExileTopWithPermission{
				From:    opp,
				GrantTo: item.Controller,
				N:       3,
				AnyType: true,
				Window:  &window,
			}.Apply(ctx)
		},
	})
}
