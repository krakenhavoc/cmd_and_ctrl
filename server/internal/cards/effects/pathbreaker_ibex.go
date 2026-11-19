package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pathbreaker Ibex — Creature — Goat {4}{G}{G}, 3/3 (Special Guests):
//
//	"Whenever this creature attacks, creatures you control gain trample and
//	 get +X/+X until end of turn, where X is the greatest power among
//	 creatures you control."
//
// Overwhelming Stampede's body on an attack trigger. X is read once as the
// trigger resolves (CR 608.2h), before any creature is pumped, so the bonus
// does not compound, and the creatures affected are fixed at that moment
// (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e52e91b-70e6-4f26-a5d4-b8e33d1debbf",
		Name:         "Pathbreaker Ibex",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pathbreaker Ibex — creatures you control gain trample and get +X/+X until end of turn, where X is the greatest power among creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := b42GreatestPowerControlledBy(g, ctx.Controller())
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{
						Match:     yours,
						Power:     x,
						Toughness: x,
						Label:     "Pathbreaker Ibex — +X/+X",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Match:    yours,
						Keywords: []string{"trample"},
						Label:    "Pathbreaker Ibex — trample",
					}.Apply(ctx)
				}),
		},
	})
}
