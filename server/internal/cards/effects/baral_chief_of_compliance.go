package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baral, Chief of Compliance — Legendary Creature — Human Wizard {1}{U},
// 1/3:
//
//	"Instant and sorcery spells you cast cost {1} less to cast.
//	 Whenever a spell or ability you control counters a spell, you may
//	 draw a card. If you do, discard a card."
//
// The reduction is Goblin Electromancer's. The trigger reads a counter that
// went through a resolving spell or ability of yours (the countered
// spell's move names the resolving item's controller as its cause, CR 701.6a),
// and only when the countered object is a spell: an ability countered
// (Stifle) draws nothing. A spell that fizzles for want of a target is not
// countered by a spell or ability. The loot is Stadium Tidalmage's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "baf80db7-2ca8-4bfc-b5b6-25ed51878dd7",
		Name:         "Baral, Chief of Compliance",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterSpell, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCounterSpell && ev.Target == ev.CardID &&
					ev.Cause == game.MoveCauseEffect && ev.CauseController == source.Controller
			}, "Baral, Chief of Compliance — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Baral, Chief of Compliance — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
