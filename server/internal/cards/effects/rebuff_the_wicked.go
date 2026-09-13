package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rebuff the Wicked — Instant {W} (EDHREC rank 1547):
//
//	"Counter target spell that targets a permanent you control."
//
// White's one-mana counterspell, narrow by design: the clause is a
// predicate over the targeted spell's own targets, read off its stack
// item — at least one of them must be a permanent the caster controls
// ON THE BATTLEFIELD right now. A spell aimed at the caster's face
// does not qualify (a player is not a permanent), and neither does
// one whose only permanent target has already left, because the
// predicate re-runs at resolution like every target clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5822e53-ddec-4c77-bcf4-091e35c1e731",
		Name:         "Rebuff the Wicked",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell that targets a permanent you control", b14SpellTargetsAPermanentYouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
