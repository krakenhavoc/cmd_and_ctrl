package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heron's Grace Champion — Creature — Human Knight {2}{G}{W}, 3/3:
//
//	"Flash
//	 Lifelink
//	 When this creature enters, other Humans you control get +1/+1 and
//	 gain lifelink until end of turn."
//
// The set of Humans is fixed as the trigger resolves (CR 611.2c) and
// excludes the Champion itself. "Human" is a word of the card's text
// (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86b5a853-9035-4490-9d0f-7c719bedfd58",
		Name:            "Heron's Grace Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heron's Grace Champion — other Humans you control get +1/+1 and gain lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					humans := effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID),
						effects.HasSubtype(g.TextReaderForEffect(item).TextSubtype("Human")))
					if err := (effects.BoostUntilEOT{Match: humans, Power: 1, Toughness: 1,
						Label: "Heron's Grace Champion — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: humans, Keywords: []string{"lifelink"},
						Label: "Heron's Grace Champion — lifelink"}.Apply(ctx)
				}),
		},
	})
}
