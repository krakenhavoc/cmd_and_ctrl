package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "86b5a853-9035-4490-9d0f-7c719bedfd58",
		Name:            "Heron's Grace Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Heron's Grace Champion — other Humans you control get +1/+1 and gain lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					humans := And(Creature(), YouControl(), OtherThan(item.SourceCardID),
						HasSubtype(g.TextReaderForEffect(item).TextSubtype("Human")))
					if err := (BoostUntilEOT{Match: humans, Power: 1, Toughness: 1,
						Label: "Heron's Grace Champion — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: humans, Keywords: []string{"lifelink"},
						Label: "Heron's Grace Champion — lifelink"}.Apply(ctx)
				}),
		},
	})
}
