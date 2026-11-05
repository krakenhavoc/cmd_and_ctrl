package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Command Subadjunct — Artifact Creature — Borg {W}, 2/1 (Star Trek,
// #4973):
//
//	"Artifact creatures you control have vigilance."
//
// A layer-6 keyword grant read on post-layer types, so an artifact made
// a creature (a crewed Vehicle, a stationed Spacecraft) has vigilance
// while it is one. The Subadjunct is an artifact creature itself and
// grants itself vigilance too.
//
// No simplification.
func init() {
	grant := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
	}, "vigilance")
	grant.Label = "Artifact creatures you control have vigilance."
	Register(Spec{
		OracleID:     "85b67121-340f-4cff-a26d-2160aa923ec3",
		Name:         "Command Subadjunct",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
