package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decimator of the Provinces — Creature — Eldrazi Boar {10}, 7/7:
//
//	"Emerge {6}{G}{G}{G} (You may cast this spell by sacrificing a
//	 creature and paying the emerge cost reduced by that creature's mana
//	 value.)
//	 When you cast this spell, creatures you control get +2/+2 and gain
//	 trample until end of turn.
//	 Trample, haste"
//
// Emerge is the shared alternative cost (ADR 0135 §4). The cast trigger
// resolves above the spell, so the creatures it pumps are the ones you
// control then: the Decimator itself is still on the stack and gets
// neither the +2/+2 nor the grant (its own trample and haste are
// printed), and a creature that arrives later this turn gets nothing (CR
// 611.2c). The pump (layer 7c) and the trample (layer 6) are Overrun's
// two primitives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "7f522ded-09fd-457e-9efe-0a6324925e4c",
		Name:             "Decimator of the Provinces",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample", "haste"},
		AlternativeCosts: []game.AlternativeCost{Emerge("{6}{G}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Decimator of the Provinces — creatures you control get +2/+2 and gain trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{
						Match:     yours,
						Power:     2,
						Toughness: 2,
						Label:     "Decimator of the Provinces — +2/+2",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Match:    yours,
						Keywords: []string{"trample"},
						Label:    "Decimator of the Provinces — trample",
					}.Apply(ctx)
				}),
		},
	})
}
