package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Split Up — Sorcery {1}{W}{W} (EDHREC rank 1961):
//
//	"Choose one —
//	 • Destroy all tapped creatures.
//	 • Destroy all untapped creatures."
//
// The one-sided wrath: cast after your opponents attacked to keep
// your own untapped team, or with vigilance to keep your attackers.
// Choose-one over two sweeps, each a DestroyAllMatching on the
// creature's tap state — tappedPermanent (The Wandering Emperor's
// predicate) and Untapped (the tap-cost one), so the two halves are
// exact complements.
//
// No simplifications. The boardwipe posture this used to declare
// (#446 — the mass-destroy path not honouring indestructible) was an
// engine gap, closed in S30 (#470).
func init() {
	Register(Spec{
		OracleID:     "2e82520a-9da3-49ae-b5c8-37e7ac8853fe",
		Name:         "Split Up",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy all tapped creatures."),
			Mode("Destroy all untapped creatures."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			sweeps := []CardPredicate{
				And(Creature(), tappedPermanent()),
				And(Creature(), Untapped()),
			}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (DestroyAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
