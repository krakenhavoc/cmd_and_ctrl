package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Too Evil to Stay Dead — {2}{B} Sorcery:
//
//	"Teamwork 4 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 4 or more.)
//	 Choose target creature card in your graveyard with mana value 4 or
//	 less. If this spell was cast using teamwork, instead choose target
//	 creature card in your graveyard. Return the chosen card to the
//	 battlefield."
//
// #1703 shipped this with the ceiling always on, because the teamwork
// cost could not carry a target-clause rewrite. #1716 lifted that
// refusal: teamwork is announced at CR 601.2b and the target at
// 601.2c, so a teamwork cast is judged — at announce, at the CR 608.2b
// re-check, in the view's picker and in the bot's enumerator — under
// the wider "target creature card in your graveyard" clause WhenPaid
// declares, and an ordinary cast under the printed one.
func init() {
	Register(Spec{
		OracleID:     "e50fecff-8872-42f5-8882-41ad13d9d1ae",
		Name:         "Too Evil to Stay Dead",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Teamwork(4),
			TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()))},
		Targets: TargetCardInGraveyard("target creature card in your graveyard with mana value 4 or less",
			YouOwn(), Creature(), ManaValueLE(4)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
		},
	})
}
