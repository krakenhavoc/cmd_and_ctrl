package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderblust — Creature — Elemental {2}{R}{R}{R}, 7/2:
//
//	"Haste
//	 This creature has trample as long as it has a -1/-1 counter on it.
//	 Persist"
//
// The trample is a layer-6 self grant read off its counters, so a
// Thunderblust that persist returned has trample and one that never
// died does not. Haste and persist are PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30d961f7-aae2-4e5a-8f83-9dc17c7dee47",
		Name:            "Thunderblust",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", game.KeywordPersist},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.Counters[game.CounterMinusOne] > 0
			}, "trample"),
		},
	})
}
