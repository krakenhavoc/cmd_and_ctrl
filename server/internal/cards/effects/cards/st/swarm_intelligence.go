package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swarm Intelligence — Enchantment {6}{U} (Commander 2021, S82 OLD b058,
// #5504):
//
//	"Whenever you cast an instant or sorcery spell, you may copy that spell.
//	 You may choose new targets for the copy."
//
// Gorion's shape on an instant or sorcery: an optional copy of the cast spell
// under this enchantment's controller, put on the stack and not cast
// (CR 707.10), with the new-targets offer (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a7a5ea9-b633-459f-bcd4-40de53afce0e",
		Name:         "Swarm Intelligence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Swarm Intelligence — you may copy that spell; you may choose new targets for the copy",
				effects.GapB106CopyTheCastSpell), "Swarm Intelligence — copy that spell?"),
		},
	})
}
