package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hancock, Ghoulish Mayor — Legendary Creature — Zombie Mutant Advisor
// {2}{B}, 2/1:
//
//	"Each other creature you control that's a Zombie or Mutant gets
//	 +X/+X, where X is the number of counters on Hancock.
//	 Undying"
//
// X counts every counter on Hancock, of every kind, read at each layer
// pass, so the +1/+1 counter undying returns it with makes X at least
// one. Undying is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "adf9aed9-dd63-48da-b799-ea94839fc22a",
		Name:            "Hancock, Ghoulish Mayor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Static: []game.StaticAbility{
			TribalScalingAnthem(TribeFilter{Tribes: []string{"Zombie", "Mutant"}, Others: true, YoursOnly: true},
				func(source *game.Card, _ *game.Game) int {
					n := 0
					for _, k := range source.Counters {
						if k > 0 {
							n += k
						}
					}
					return n
				}),
		},
	})
}
