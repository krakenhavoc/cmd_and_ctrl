package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earthshaker Giant — Creature — Giant Druid {4}{G}{G}, 6/6 (Game Night
// 2019, #5515):
//
//	"Trample
//	 When this creature enters, other creatures you control get +3/+3 and
//	 gain trample until end of turn."
//
// Rumbleweed's enters trigger without its cost reduction: the creatures
// you control other than the Giant as the trigger resolves (CR 611.2c),
// so a creature that arrives later in the turn is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cd6250ae-9079-4a62-8a70-0d94fbac21bc",
		Name:            "Earthshaker Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Earthshaker Giant — other creatures you control get +3/+3 and gain trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					self := ctx.Source()
					match := And(Creature(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
					if err := (BoostUntilEOT{Match: match, Power: 3, Toughness: 3, Label: "Earthshaker Giant — +3/+3"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"}, Label: "Earthshaker Giant — trample until end of turn"}.Apply(ctx)
				}),
		},
	})
}
