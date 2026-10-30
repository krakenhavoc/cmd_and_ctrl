package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d2ba20a0-7c08-49ed-b960-4b3154a8c025",
		Name:         "Iron Suitcase",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Iron Suitcase — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}: This artifact becomes a 3/3 Construct artifact creature with flying until end of turn.",
			Cost:    ManaCost("{3}"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if sourceIsNewObject(g, item) {
					return nil
				}
				if err := (BecomeCreatureUntilEOT{
					Subtypes: []string{"Construct"}, SetPower: 3, SetToughness: 3,
					Label: "Iron Suitcase — 3/3 Construct artifact creature",
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Iron Suitcase — flying"}.Apply(ctx)
			},
		}},
	})
}
