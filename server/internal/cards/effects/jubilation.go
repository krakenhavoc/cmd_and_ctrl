package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jubilation — Creature — Elemental Incarnation {5}{G}, 5/5 (Lorwyn
// Eclipsed Commander):
//
//	"When this creature enters, creatures you control get +2/+2 and gain
//	 trample until end of turn.
//	 Encore {7}{G}{G} (...)"
//
// The set of creatures is pinned as the ability resolves (CR 611.2c): one
// that enters afterwards gets nothing. Encore is the shared keyword
// (encore.go); each hasty token copy pumps the team when it enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2cf04076-adc4-4831-a3b8-9120d516f204",
		Name:         "Jubilation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jubilation — creatures you control get +2/+2 and gain trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					match := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: match, Power: 2, Toughness: 2, Label: "Jubilation — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"}, Label: "Jubilation — trample"}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{Encore("{7}{G}{G}")},
	})
}
