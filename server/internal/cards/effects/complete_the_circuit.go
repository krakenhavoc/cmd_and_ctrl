package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Complete the Circuit — Instant {5}{U} (March of the Machine):
//
//	"Convoke
//	 You may cast sorcery spells this turn as though they had flash.
//	 When you next cast an instant or sorcery spell this turn, copy that
//	 spell twice. You may choose new targets for the copies."
//
// Convoke is Spec.TapCost. The first sentence is the cast-timing grant
// Alchemist's Refuge makes, narrowed to sorceries. The second is a delayed
// trigger that waits for the next instant or sorcery cast this turn
// (Doublecast), and copies it twice with CR 707.10c's new-targets offer
// (Summon: G.F. Cerberus's copy body). It does not copy Complete the
// Circuit itself: that cast was over before this resolved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "402506fa-6bd5-4814-8353-90c148dd576c",
		Name:         "Complete the Circuit",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GrantCastTiming{
				Filter: game.PermissionFilter{SorceryOnly: true},
				Label:  "You may cast sorcery spells this turn as though they had flash.",
			}).Apply(ctx); err != nil {
				return err
			}
			return WhenYouNextCast(
				"Complete the Circuit — copy that spell twice",
				game.CastFilter{Types: []string{"Instant", "Sorcery"}},
				copyTheSpellTwiceBody,
			).Apply(ctx)
		},
	})
}
