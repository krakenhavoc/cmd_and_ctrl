package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gallia of the Endless Dance — Legendary Creature — Satyr {R}{G}, 2/2 (Theros
// Beyond Death, S82 OLD, #5514):
//
//	"Haste
//	 Other Satyrs you control get +1/+1 and have haste.
//	 Whenever you attack with three or more creatures, you may discard a card
//	 at random. If you do, draw two cards."
//
// The lord is a tribal anthem and a keyword grant over the other Satyrs its
// controller controls (a changeling counts). "Whenever you attack with three
// or more creatures" is Warcry Phoenix's batch trigger, one per declaration
// of attackers with at least three. The "you may" is asked as it resolves;
// the discard is at random from the hand (CR 701.9b), and the two cards are
// drawn only if a card was really discarded.
//
// No simplification.
func init() {
	filter := effects.TribeFilter{Tribes: []string{"Satyr"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "43f135b5-d33e-4873-bc5a-8c939582d624",
		Name:            "Gallia of the Endless Dance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(filter, 1, 1),
			effects.TribalKeywordGrant(filter, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Gallia of the Endless Dance — you may discard a card at random; if you do, draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return g.DiscardRandomKeyedForEffect(item.Controller, 1, effects.NewContext(g, item).ItemTail(galliaDrawTwo))
				}),
				effects.ThatManyEventsAtLeast(3)),
				"Gallia of the Endless Dance — discard a card at random? (If you do, draw two cards.)"),
		},
	})
}

// galliaDrawTwo is "if you do, draw two cards": only a card really discarded
// draws. An on-disk key.
var galliaDrawTwo = effects.CardTail("gallia-of-the-endless-dance/draw-two", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
})
