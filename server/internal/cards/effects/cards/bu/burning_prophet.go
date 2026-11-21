package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Prophet — Creature — Human Wizard {1}{R}, 1/3 (Ravnica Remastered,
// #5948):
//
//	"Whenever you cast a noncreature spell, this creature gets +1/+0 until
//	 end of turn, then scry 1."
//
// Black Bolt's cast trigger, with Artificer's Assistant's scry after the pump.
// The pump lands only while the Prophet is still the permanent the ability came
// from (CR 400.7); the scry is the controller's and happens either way ("then",
// not "if you do"). The Prophet's own cast does not trigger it: it is a
// creature spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e462ce2-15d3-473c-8f48-9c9363fc65eb",
		Name:         "Burning Prophet",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Burning Prophet — it gets +1/+0 until end of turn, then scry 1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.SourceIsStillThisPermanent(g, item) {
						if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Burning Prophet — +1/+0"}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.Scry{N: 1}.Apply(ctx)
				}),
		},
	})
}
