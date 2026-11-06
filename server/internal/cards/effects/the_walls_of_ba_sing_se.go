package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Walls of Ba Sing Se — Legendary Artifact Creature — Wall {8}, 0/30:
//
//	"Defender
//	 Other permanents you control have indestructible."
//
// Defender is printed. The second line is a layer-6 grant (CR 613.1f) to
// every other permanent its controller controls, lands and artifacts as
// much as creatures, the same static as Privileged Position's with
// indestructible for hexproof. It is not locked in (CR 611.3a): a
// permanent that arrives later has it at once, and the Walls leaving
// takes it away, so damage already marked can become lethal that turn
// (the ruling).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "af7d7da5-aff5-48e1-901a-f81257a85f71",
		Name:            "The Walls of Ba Sing Se",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller
			}, "indestructible"),
		},
	})
}
