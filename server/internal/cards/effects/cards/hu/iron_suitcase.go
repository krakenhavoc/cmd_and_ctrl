package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Suitcase — {1} Artifact (Marvel Super Heroes Commander, #4455):
//
//	"When this artifact enters, scry 2.
//	 {3}: This artifact becomes a 3/3 Construct artifact creature with flying
//	 until end of turn."
//
// The animation is BecomeCreatureUntilEOT (types, subtype and base 3/3) plus
// a flying grant, both for the turn. It is no Vehicle: it can be activated
// at any time and keeps its other characteristics.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2ba20a0-7c08-49ed-b960-4b3154a8c025",
		Name:         "Iron Suitcase",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Iron Suitcase — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This artifact becomes a 3/3 Construct artifact creature with flying until end of turn.",
			Cost:    effects.ManaCost("{3}"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				if err := (effects.BecomeCreatureUntilEOT{
					Subtypes: []string{"Construct"}, SetPower: 3, SetToughness: 3,
					Label: "Iron Suitcase — 3/3 Construct artifact creature",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Iron Suitcase — flying"}.Apply(ctx)
			},
		}},
	})
}
