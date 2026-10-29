package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thorin Oakenshield — Legendary Creature — Dwarf Noble {R}{W}, 3/2:
//
//	"Trample
//	 Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 As long as you have an enduring story, artifacts and creatures you
//	 control have ward {1}."
//
// Storied is the engine's (CR 702.195a). The ward is Flowering of the
// White Tree's granted ward (WardGranted): Thorin carries the trigger for
// each artifact and creature you control, Thorin included, and it asks
// for the story as the permanent becomes a target.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:        "bdd41af0-bbd1-4ecd-a699-99f006f5e5ce",
		Name:            "Thorin Oakenshield",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordStoried},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Thorin Oakenshield — artifacts and creatures you control have ward {1}",
				WhileEnduringStory(func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && (target.IsArtifact() || target.IsCreature())
				})),
		},
	})
}
