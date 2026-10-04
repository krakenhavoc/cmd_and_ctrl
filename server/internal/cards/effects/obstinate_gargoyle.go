package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obstinate Gargoyle — Artifact Creature — Gargoyle {1}{W}{B}, 2/2:
//
//	"This creature has flying as long as it's modified. (Equipment,
//	 Auras you control, and counters are modifications.)
//	 Persist"
//
// The -1/-1 counter persist returns it with makes it modified, so a
// returned Gargoyle flies. Persist is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f3bd9bfe-847f-4ed1-88aa-6aa6b3fa7199",
		Name:            "Obstinate Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && b07IsModified(g, *target)
			}, "flying"),
		},
	})
}
