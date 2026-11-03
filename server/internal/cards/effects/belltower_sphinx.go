package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Belltower Sphinx — Creature — Sphinx {4}{U}, 2/5:
//
//	"Flying
//	 Whenever a source deals damage to this creature, that source's
//	 controller mills that many cards."
//
// Phyrexian Obliterator's trigger with a mill for the sacrifice: one
// trigger per damage event (the engine emits EventDealDamage per source,
// with the damaged card in Target and the amount in Amount), so two
// blockers are two triggers. The source's controller and the amount are
// read as the trigger fires and carried on the stack item, so the mill
// still happens if the source or the Sphinx has left by resolution. A
// library holding fewer cards mills what it has (CR 701.17b).
//
// No simplification.
func init() {
	const key = "Belltower Sphinx — that source's controller mills that many cards"
	Register(Spec{
		OracleID:        "b9ed0081-7eda-4ff9-ac2d-bcdc29a22d9f",
		Name:            "Belltower Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b17SourceDealtDamageToSelf(ev, source)
			},
			Key: key,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, key)
				item.Params.Player = b17DamageSourceController(ev, g)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return MillCards{Player: item.Params.Player, N: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
