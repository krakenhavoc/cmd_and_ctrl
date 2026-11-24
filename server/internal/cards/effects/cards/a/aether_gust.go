package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Gust — Instant {1}{U} (Secret Lair Drop):
//
//	"Choose target spell or permanent that's red or green. Its owner puts it on
//	 their choice of the top or bottom of their library."
//
// One target clause that is either a spell on the stack or a permanent on the
// battlefield (Brutal Expulsion's TargetSpellOrPermanent), either way red or green
// as the layers have it. As it resolves the OWNER of the object picks the lane
// (Run Ashore's PutIntoLibraryAtDepthOrBottom at depth 1, the top), and a
// commander's owner is offered the command zone (CR 903.9). A spell put into a
// library is not countered, and a target gone by resolution leaves nothing to do
// (CR 608.2b).
//
// No simplification.
func init() {
	redOrGreen := effects.Or(effects.OfColor("R"), effects.OfColor("G"))
	effects.Register(effects.Spec{
		OracleID:     "ecf1e378-d6bc-4ee0-800e-eb8a955d781f",
		Name:         "Aether Gust",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpellOrPermanent("target spell or permanent that's red or green", redOrGreen, redOrGreen),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.PutIntoLibraryAtDepthOrBottom{Card: t.ID, Depth: 1,
					Label: "Aether Gust — put it on your choice of the top or bottom of your library"}.Apply(ctx)
			}
			return nil
		},
	})
}
