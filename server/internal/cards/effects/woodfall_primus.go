package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodfall Primus — Creature — Treefolk Shaman {5}{G}{G}{G}, 6/6:
//
//	"Trample
//	 When this creature enters, destroy target noncreature permanent.
//	 Persist"
//
// Trample and persist are PrintedKeywords (#2075); the return is a
// second entry, so it destroys a second permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f70f1bb-29fa-4abb-afc2-653acd0a08b9",
		Name:            "Woodfall Primus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordPersist},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Woodfall Primus — destroy target noncreature permanent",
				func(g *game.Game, item *game.StackItem) error {
					return destroyTheTargetPermanent(item, NewContext(g, item))
				}), TargetPermanent("target noncreature permanent", Noncreature())),
		},
	})
}
