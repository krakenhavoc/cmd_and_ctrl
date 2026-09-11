package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Ritual — Instant {B} (EDHREC rank 33):
//
//	"Add {B}{B}{B}."
//
// The original fast mana, and the highest-ranked card in the
// roadmap's "mana pipeline" group: a SPELL that adds mana. It could
// never be a ManaAbility — mana abilities don't use the stack (CR
// 605.3a) and this card does, which is why it can be countered and
// why Storm-Kiln Artist triggers on it — so until the AddMana
// primitive existed there was no way to write it at all.
//
// The three black mana land in the pool as the spell resolves and
// empty with the pool at the end of the step (CR 106.4): cast it
// in the main phase with something to spend it on, as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "53f7c868-b03e-4fc2-8dcf-a75bbfa3272b",
		Name:     "Dark Ritual",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return AddMana{Produced: "{B}{B}{B}"}.Apply(ctx)
		},
	})
}
