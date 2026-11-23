package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voice of Many — Creature — Elf Druid {2}{G}{G}, 3/3 (Commander 2019, #5517):
//
//	"When this creature enters, draw a card for each opponent who controls
//	 fewer creatures than you."
//
// The count is taken as the trigger resolves (CR 608.2h): every opponent
// still in the game who controls fewer creatures than you do then, the Voice
// itself counted among yours if it is still there. No opponent behind, no
// draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b40faad-1b96-4461-89d8-9a207b58e832",
		Name:         "Voice of Many",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Voice of Many — draw a card for each opponent who controls fewer creatures than you",
				c19VoiceOfManyDraw),
		},
	})
}

// c19VoiceOfManyDraw draws one card per opponent with fewer creatures than
// the trigger's controller.
func c19VoiceOfManyDraw(g *game.Game, item *game.StackItem) error {
	mine := effects.B04CreaturesControlled(g, item.Controller)
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == item.Controller {
			continue
		}
		if effects.B04CreaturesControlled(g, p.ID) < mine {
			n++
		}
	}
	return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
}
