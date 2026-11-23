package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// On the Shoulders of Giants — {2}{U} Enchantment (Star Trek Commander,
// #4974):
//
//	"Historic creatures you control have flying. (Artifacts,
//	 legendaries, and Sagas are historic.)"
//
// A layer-6 keyword grant (CR 613.1f) over the creatures you control
// that are historic (CR 700.6) as the layers see them: an artifact, a
// legendary permanent or a Saga.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8abb1be2-d02b-4a78-9fcb-d2c9102a840b",
		Name:         "On the Shoulders of Giants",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && effects.B09IsHistoric(*target)
			}, "flying"), "Historic creatures you control have flying."),
		},
	})
}
