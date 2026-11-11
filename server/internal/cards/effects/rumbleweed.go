package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rumbleweed:
//
//	"This spell costs {1} less to cast for each land card in your graveyard.
//	 Reach, vigilance, trample
//	 When this creature enters, other creatures you control get +3/+3 and
//	 gain trample until end of turn."
//
// The discount is counted as the cost is determined (CR 601.2f) and
// reduces only the generic part; the mana value stays 11. The enters
// trigger affects the creatures you control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9cfc01a7-e291-4388-bd09-94846d72e291",
		Name:            "Rumbleweed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(cmdr2S14LandCardsInYourGraveyard,
				"This spell costs {1} less to cast for each land card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rumbleweed — other creatures you control get +3/+3 and gain trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					self := ctx.Source()
					match := And(Creature(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
					if err := (BoostUntilEOT{Match: match, Power: 3, Toughness: 3, Label: "Rumbleweed — +3/+3"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"}, Label: "Rumbleweed — trample until end of turn"}.Apply(ctx)
				}),
		},
	})
}
