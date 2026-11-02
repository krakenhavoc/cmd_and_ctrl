package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Privileged Position — Enchantment {2}{G/W}{G/W}{G/W}:
//
//	"({G/W} can be paid with either {G} or {W}.)
//	 Other permanents you control have hexproof."
//
// A static ability that adds hexproof (layer 6, CR 613.1f) to every
// permanent the source's controller controls except the enchantment
// itself: lands, artifacts and planeswalkers as well as creatures. It is
// not locked in (CR 611.3a), so a permanent that arrives later has hexproof
// at once and one that changes hands loses it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abd62af0-c17d-4f62-af15-9ea83037b990",
		Name:         "Privileged Position",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller
			}, "hexproof"),
		},
	})
}
