package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lashweed Lurker — Creature — Eldrazi Horror {8}, 5/4:
//
//	"Emerge {5}{G}{U} (You may cast this spell by sacrificing a creature
//	 and paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, you may put target nonland permanent on
//	 top of its owner's library."
//
// Emerge is the shared alternative cost (ADR 0135 §4). The cast trigger
// is a "you may", asked as it triggers; on a yes the nonland permanent
// goes on top of its owner's library above the spell, even if the Lurker
// is countered. A target gone by then is left alone (CR 608.2b), and a
// token put into a library ceases to exist (CR 111.7).
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Lashweed Lurker — put target nonland permanent on top of its owner's library",
		func(g *game.Game, item *game.StackItem) error {
			return putTargetOnTopOfOwnersLibrary(item, NewContext(g, item))
		})
	cast.Targets = TargetPermanent("target nonland permanent", Nonland())
	Register(Spec{
		OracleID:         "a7dd75bf-ecda-406c-baa2-8c051f138809",
		Name:             "Lashweed Lurker",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Emerge("{5}{G}{U}")},
		Triggered: []game.TriggeredAbility{
			Optional(cast, "Lashweed Lurker — put a nonland permanent on top of its owner's library?"),
		},
	})
}
