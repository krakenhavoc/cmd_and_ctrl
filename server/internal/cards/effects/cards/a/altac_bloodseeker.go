package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Altac Bloodseeker — Creature — Human Berserker {1}{R}, 2/1 (Magic 2015,
// #5556):
//
//	"Whenever a creature an opponent controls dies, this creature gets +2/+0
//	 and gains first strike and haste until end of turn."
//
// Assault Intercessor's watch (a creature that died under an opponent's
// control as it last existed, CR 603.10a) with a pump and two keyword grants
// on the Bloodseeker itself. A Bloodseeker that has left, or come back as a
// new object, is not changed (CR 400.7). It sees its own controller's
// creatures die as nothing.
//
// No simplification.
func init() {
	const label = "Altac Bloodseeker — gets +2/+0 and gains first strike and haste until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "7f9d5219-0306-4c39-8640-dcb4314042de",
		Name:         "Altac Bloodseeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) != source.Controller
			},
			Key: label,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, label)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Altac Bloodseeker — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike", "haste"}, Label: "Altac Bloodseeker — first strike and haste"}.Apply(ctx)
			},
		}},
	})
}
