package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watchful Radstag — Creature — Elk Mutant {2}{G}, 2/2:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Whenever this creature evolves, create a token that's a copy of
//	 it."
//
// The second "evolves" card (#1805, ADR 0106 §3): it watches
// EventEvolved (CR 702.100b) through WhenThisEvolves.
//
// The token copies the Radstag's copiable values (CR 707.2), so it is a
// 2/2 with evolve and this same trigger, and its counters are not
// copied. It enters as a creature under its controller's control, so
// it triggers the original's evolve and every other evolve the
// controller has — the Radstag's own evolve only if the token is
// bigger, which a copy of it never is. Scute Swarm's shape
// (CreateTokenCopy of the source).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "572afa0f-9e82-4899-b622-a7554625b9f4",
		Name:            "Watchful Radstag",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Triggered: []game.TriggeredAbility{
			WhenThisEvolves("Watchful Radstag — create a token that's a copy of it", func(g *game.Game, item *game.StackItem) error {
				return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
