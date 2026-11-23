package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baloth Woodcrasher — Creature — Beast {4}{G}{G}, 4/4 (Foundations
// Jumpstart):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +4/+4 and gains trample until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "faaf9975-74df-4e9d-be0c-19a446cd507c",
		Name:         "Baloth Woodcrasher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Baloth Woodcrasher — landfall: +4/+4 and trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 4, Toughness: 4,
						Label: "Baloth Woodcrasher — +4/+4"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"},
						Label: "Baloth Woodcrasher — trample"}.Apply(ctx)
				}),
		},
	})
}
