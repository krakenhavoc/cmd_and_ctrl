package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scheming Symmetry — Sorcery {B} (Secret Lair Drop):
//
//	"Choose two target players. Each of them searches their library for a card,
//	 then shuffles and puts that card on top."
//
// The two players are one clause of two distinct picks (The Brothers' War's
// "two target players"), the caster included if chosen. Each target in turn
// order searches their own library for any card, shuffles it, and puts the found
// card on top (the one-mana tutor's ToTop, CR 701.24). A target that is no longer
// legal as it resolves is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dfb50d7-359b-4644-90e3-7ecdacee7c12",
		Name:         "Scheming Symmetry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetPlayer("two target players").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if err := (effects.SearchLibrary{
					Player:  t.ID,
					Limit:   1,
					Dest:    game.ZoneLibrary,
					ToTop:   true,
					Shuffle: true,
					Reason:  "Scheming Symmetry — search your library for a card",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
