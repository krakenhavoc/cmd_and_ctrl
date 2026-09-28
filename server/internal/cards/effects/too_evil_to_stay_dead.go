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
// #1703: Teamwork(4) is straightforward; the target clause is not.
// TargetSpec.CardOK is a static predicate over (game, caster,
// candidate) — it has no way to read "was teamwork announced WITH
// THIS CAST", which lives on the stack item the predicate is never
// handed (unlike TargetSpec.ManaValueAtMostX, which is an engine flag
// resolved from the announcement for exactly this reason, and which
// only widens an X-bound, not an optional-cost-bound, clause). Wiring
// a target predicate to an optional cost's announcement is new
// targeting machinery, out of scope here (see AGENTS.md's "when NOT to
// add a catalog entry" — the cost has a shape, but this consequence of
// it does not yet).
//
// So the clause always reads "mana value 4 or less": a teamwork cast
// does not get the wider pool the oracle text promises. Weaker than
// printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "e50fecff-8872-42f5-8882-41ad13d9d1ae",
		Name:         "Too Evil to Stay Dead",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Casting with teamwork doesn't widen the target — you can only choose a creature card with mana value 4 or less from your graveyard, even when the teamwork cost is paid.",
		},
		OptionalCosts: []game.AdditionalCost{Teamwork(4)},
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
