package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lucky Clover — Artifact {2} (Throne of Eldraine, S82 old sets):
//
//	"Whenever you cast an Adventure instant or sorcery spell, copy it. You may
//	 choose new targets for the copy."
//
// Gorion's trigger without the "may": an Adventure spell is a spell whose
// showing half is an Adventure, which is always an instant or a sorcery, so the
// creature half of an adventurer card is not one. The copy is of the spell the
// event names, put on the stack under the Clover's controller, not cast (CR
// 707.10), with new targets allowed (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "370d67f6-8d43-46ca-ae6c-00799bb0bd0c",
		Name:         "Lucky Clover",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.AdventureSpellCast()),
				"Lucky Clover — copy that Adventure spell; you may choose new targets for the copy",
				effects.GapB106CopyTheCastSpell),
		},
	})
}
