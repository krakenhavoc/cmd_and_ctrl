package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kruin Striker — Creature — Human Warrior {1}{R}, 2/1:
//
//	"Whenever another creature you control enters, this creature gets
//	 +1/+0 and gains trample until end of turn."
//
// Kinsbaile Aspirant's trigger with a power boost and a turn-scoped
// trample grant.
//
// No simplification.
func init() {
	const label = "Kruin Striker — +1/+0 and trample until end of turn"
	Register(Spec{
		OracleID:     "3e261fa8-1256-4525-8115-5880758338fe",
		Name:         "Kruin Striker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"trample"}, Label: label}.Apply(ctx)
			}),
		},
	})
}
