package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stagecoach Security — Creature — Human Soldier {4}{W}, 4/5 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, creatures you control get +1/+1 and gain
//	 vigilance until end of turn.
//	 Plot {3}{W} (You may pay {3}{W} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// The enters trigger is The Super Hero Civil War's chapter II: the
// affected set is the creatures you control as it resolves, the
// Security included (CR 611.2c), so a creature that arrives later in the
// turn gets neither. Plot is the ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "4b4589ff-d141-49a4-a0bd-1c73fc3651ef",
		Name:           "Stagecoach Security",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{3}{W}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stagecoach Security — creatures you control get +1/+1 and gain vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Stagecoach Security — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"vigilance"}, Label: "Stagecoach Security — vigilance"}.Apply(ctx)
				}),
		},
	})
}
