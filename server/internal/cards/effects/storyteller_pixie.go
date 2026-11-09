package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storyteller Pixie — Creature — Faerie Wizard {3}{U}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Whenever you cast an Adventure spell, draw a card."
//
// An "Adventure spell" is a spell whose showing half is an Adventure (its
// type line says so), so the creature half of an adventurer card, cast
// from hand or from exile, is not one. The trigger fires as the spell is
// cast (CR 603.2), before it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51bb6d3b-be6a-435d-8a80-0c0f0161892d",
		Name:            "Storyteller Pixie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(AdventureSpellCast()), "Storyteller Pixie — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}
