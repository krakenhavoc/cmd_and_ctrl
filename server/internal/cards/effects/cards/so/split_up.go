package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2e82520a-9da3-49ae-b5c8-37e7ac8853fe",
		Name:         "Split Up",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Destroy all tapped creatures."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
			effects.ModeWithPurpose(effects.Mode("Destroy all untapped creatures."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			sweeps := []effects.CardPredicate{
				effects.And(effects.Creature(), effects.TappedPermanent()),
				effects.And(effects.Creature(), effects.Untapped()),
			}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (effects.DestroyAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
