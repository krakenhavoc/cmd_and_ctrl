package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Root Spider — Creature — Spider {3}{G}, 2/2 (Homelands, S82 OLD b823, #5635):
//
//	"Whenever this creature blocks, it gets +1/+0 and gains first strike until end
//	of turn."
//
// Giant Badger's block trigger with a pump and a keyword grant, both until the
// cleanup step (CR 514.2) and both from the one trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8fe0115-5d98-4251-a48c-5e849430a445",
		Name:         "Root Spider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Root Spider — gets +1/+0 and gains first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Root Spider — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"}, Label: "Root Spider — first strike"}.Apply(ctx)
				}),
		},
	})
}
