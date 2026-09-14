package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Singularity Rupture — Sorcery {3}{U}{B}{B} (EDHREC rank 2387):
//
//	"Destroy all creatures, then any number of target players each
//	 mill half their library, rounded down."
//
// A Damnation with a mill rider. The wipe is the shared
// DestroyAllMatching primitive; the mill is one MillCards per legal
// target, half of that player's library as it stands when the spell
// resolves (after the wipe — "then"), rounded down. "Any number of
// target players" is a zero-to-unbounded player clause, so the
// caster may name nobody, everyone, or just the graveyard deck.
//
// No simplifications. The caveat this used to carry — the
// mass-destroy path not honouring indestructible (#446) — was an
// engine gap, closed in S30 (#470).
func init() {
	Register(Spec{
		OracleID:     "e1976b6c-7e43-4f4a-b082-f95495a1b260",
		Name:         "Singularity Rupture",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("any number of target players").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DestroyAllMatching{Match: Creature()}).Apply(ctx); err != nil {
				return err
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if err := b22MillHalf(ctx, t.ID); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
