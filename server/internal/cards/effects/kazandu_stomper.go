package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kazandu Stomper — Creature — Beast {5}{G}, 6/5:
//
//	"Trample
//	 When this creature enters, return up to two lands you control to
//	 their owner's hand."
//
// The lands are chosen as the trigger resolves (zero, one or two of the
// controller's lands, Bull Elephant's pick shape without the sacrifice)
// and go back in one return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "121f6c96-2911-4ee3-b14e-d3b143eb14c3",
		Name:            "Kazandu Stomper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kazandu Stomper — return up to two lands you control to their owner's hand", kazanduStomperEnters),
		},
	})
}

var kazanduStomperPick = askedPick("choose-cards/kazandu-stomper-return-lands", func(ctx *Context, c game.CardsChosen) error {
	live := c.Picked[:0:0]
	for _, id := range c.Picked {
		if onBattlefield(ctx.Game, id) {
			live = append(live, id)
		}
	}
	if len(live) == 0 {
		return nil
	}
	ctx.Game.BounceCardsToHandForEffect(live)
	return nil
})

func kazanduStomperEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var lands []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsLand() {
			lands = append(lands, c.InstanceID)
		}
	}
	if len(lands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   "Kazandu Stomper — return up to two lands you control to their owner's hand",
		Cards:      lands,
		Min:        0,
		Max:        2,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    kazanduStomperPick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}
