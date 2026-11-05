package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primary Unimatrix Adjunct — Artifact Creature — Borg {3}{W}, 2/3
// (Star Trek, #4973):
//
//	"Artifact creatures you control have double strike."
//
// A layer-6 keyword grant (KeywordGrant) read in the layer pass, so an
// artifact animated in layer 4 is an artifact creature by the time
// layer 6 asks (Cyberdrive Awakener's reading). The Adjunct is an
// artifact creature itself, so it has double strike too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be3dd0c8-32c2-4a56-9290-1fc9e366bae2",
		Name:         "Primary Unimatrix Adjunct",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsArtifact() && target.IsCreature()
			}, "double strike"), "Artifact creatures you control have double strike."),
		},
	})
}
