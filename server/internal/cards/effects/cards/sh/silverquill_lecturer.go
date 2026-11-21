package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverquill Lecturer — Creature — Kor Wizard {4}{W}, 3/3 (Modern Horizons
// 3 Commander, #5948):
//
//	"Creature spells you cast have demonstrate. (Whenever you cast a
//	 creature spell, you may copy it. If you do, choose an opponent to also
//	 copy it. Each copy becomes a token.)"
//
// The other shape of demonstrate (Prismari's storm, Maelstrom Nexus's
// cascade): a permanent that GIVES the keyword, so the trigger lives on
// the battlefield and watches the cast of a creature spell its controller
// casts. Each such cast offers the "you may" as the trigger resolves
// (CR 702.144a); a spell that also prints demonstrate triggers a second
// time. The copies are tokens (CR 707.10f) and are not cast, so they do
// not demonstrate again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac230b68-f945-4a1a-bd89-33e6d62e3c36",
		Name:         "Silverquill Lecturer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GapB45GrantsDemonstrate("Silverquill Lecturer", func(spell game.Card, _ *game.Card, _ *game.Game) bool {
				return spell.IsCreature()
			}),
		},
	})
}
