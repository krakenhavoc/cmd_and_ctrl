package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kathari Remnant — Creature — Bird Skeleton {2}{U}{B}, 0/1 (Alara
// Reborn, #5586):
//
//	"Flying
//	 {B}: Regenerate this creature.
//	 Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast
//	 it without paying its mana cost. Put the exiled cards on the bottom in
//	 a random order.)"
//
// Cadaverous Knight's regeneration row at {B}, and the catalog cascade
// keyword (CR 702.85a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bbf07946-33d0-48fe-85af-022e9b0b834c",
		Name:            "Kathari Remnant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.Cascade()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
