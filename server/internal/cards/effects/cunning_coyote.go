package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cunning Coyote — Creature — Coyote {1}{R}, 2/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Haste
//	 When this creature enters, another target creature you control gets
//	 +1/+1 and gains haste until end of turn.
//	 Plot {1}{R} (You may pay {1}{R} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// "Another" is object identity (Another), so the Coyote can't pick
// itself. The +1/+1 is a layer 7c pump and the haste a layer 6 grant,
// both until cleanup, on the still-legal target (CR 608.2b). Plot is the
// ordinary special action (CR 702.170); a plotted Coyote cast for free
// still enters and still triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a38eba6a-db90-48ac-bb49-889a7fd90ce9",
		Name:            "Cunning Coyote",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		SpecialActions:  []game.SpecialAction{Plot("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Cunning Coyote — another target creature you control gets +1/+1 and gains haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Cunning Coyote — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Cunning Coyote — haste"}.Apply(ctx)
				}),
				Another(TargetCreature("another target creature you control", YouControl()))),
				ForTargets(BenefitUntilEOT(0, 1, 1, "haste"))),
		},
	})
}
