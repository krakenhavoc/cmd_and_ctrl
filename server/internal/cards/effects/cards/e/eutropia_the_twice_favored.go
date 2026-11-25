package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eutropia the Twice-Favored — Legendary Creature — Human Wizard {1}{G}{U}, 2/2:
//
//	"Constellation — Whenever an enchantment you control enters, put a +1/+1
//	 counter on target creature. That creature gains flying until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4de1319f-b3c7-42d4-9342-3e9c4f526ee0",
		Name:         "Eutropia the Twice-Favored",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.GapB103EnchantmentEnteredUnderYourControl,
					"Eutropia the Twice-Favored — put a +1/+1 counter on target creature; it gains flying until end of turn (constellation)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"},
							Label: "Eutropia the Twice-Favored — flying until end of turn"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
