package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8abb1be2-d02b-4a78-9fcb-d2c9102a840b",
		Name:         "On the Shoulders of Giants",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && b09IsHistoric(*target)
			}, "flying"), "Historic creatures you control have flying."),
		},
	})
}
