package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorion, Wise Mentor — Legendary Creature — Human Wizard {G}{W}{U}, 3/4
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Vigilance
//	 Whenever you cast an Adventure spell, you may copy it. You may choose new
//	 targets for the copy."
//
// An "Adventure spell" is a spell whose showing half is an Adventure (Storyteller
// Pixie's read), so the creature half of an adventurer card is not one. The
// trigger is Codie's copy of the cast spell: an optional (CR 603.5) copy of the
// spell the event names, put on the stack under Gorion's controller, not cast
// (CR 707.10), with new targets allowed (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "195bd8fe-581c-44ed-b7f8-e800df3502ff",
		Name:            "Gorion, Wise Mentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.YouCast(effects.AdventureSpellCast()),
				"Gorion, Wise Mentor — you may copy that Adventure spell; you may choose new targets for the copy",
				effects.GapB106CopyTheCastSpell), "Gorion, Wise Mentor — copy that Adventure spell?"),
		},
	})
}
