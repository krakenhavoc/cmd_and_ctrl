package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3e261fa8-1256-4525-8115-5880758338fe",
		Name:         "Kruin Striker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"trample"}, Label: label}.Apply(ctx)
			}),
		},
	})
}
