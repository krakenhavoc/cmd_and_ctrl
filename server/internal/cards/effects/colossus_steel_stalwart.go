package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossus, Steel Stalwart — Legendary Creature — Mutant Hero {4}{G},
// 4/5 (Marvel Super Heroes Commander, #4455):
//
//	"During your turn, Colossus has indestructible.
//	 Other Mutants you control get +1/+1."
//
// The indestructible grant is Cloud, Planet's Champion's shape: a
// keyword grant that applies to the source alone and only while it is
// its controller's turn, re-read on every layer recompute. The anthem
// is the shared tribal one over OTHER Mutants the controller controls.
//
// No simplification.
func init() {
	indestructible := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && IsYourTurn(g, source.Controller)
	}, "indestructible")
	indestructible.Label = "During your turn, Colossus has indestructible."
	Register(Spec{
		OracleID:     "fb47ceb6-b3b3-4e81-9c15-b9990c85e892",
		Name:         "Colossus, Steel Stalwart",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			indestructible,
			TribalAnthem(TribeFilter{Tribes: []string{"Mutant"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
