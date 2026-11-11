package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Villainous Wealth — Sorcery {X}{B}{G}{U} (Outlaws of Thunder Junction
// Commander, #5368):
//
//	"Target opponent exiles the top X cards of their library. You may
//	 cast any number of spells with mana value X or less from among
//	 them without paying their mana costs."
//
// Kotis, the Fangkeeper's seam with an announced X and a target
// (ExileTopWithPermission.EachManaValueAtMost): the top X cards of the
// target opponent's library are exiled face up, with one free,
// pass-closed permission held by the caster over the cards that landed,
// capped at mana value X on each spell with no running total. The casts
// are made during the resolution (CR 608.2g), each its own answer until
// the caster declines or none is left; a card not cast, a land and a
// spell over mana value X stay in exile. An {X} in a spell's cost is 0
// (the free cast locks it). X is the announced X (CR 601.2b); an X of
// zero exiles nothing. If the target opponent is an illegal target by
// resolution, the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2fae95d-074b-4c0c-99cd-f9385a424c05",
		Name:         "Villainous Wealth",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			limit := ctx.X()
			if limit <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return ExileTopWithPermission{
					From:                t.ID,
					GrantTo:             item.Controller,
					N:                   limit,
					EachManaValueAtMost: &limit,
				}.Apply(ctx)
			}
			return nil
		},
	})
}
