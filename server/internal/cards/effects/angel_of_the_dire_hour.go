package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of the Dire Hour — Creature — Angel {5}{W}{W}, 5/4 (Jumpstart
// Selesnya deck request, #4893):
//
//	"Flash
//	 Flying
//	 When this creature enters, if you cast it from your hand, exile all
//	 attacking creatures."
//
// The Settle the Wreckage on a flash body. "If you cast it from your hand"
// is an intervening if (CR 603.4), Wakening Sun's Avatar's read
// (b30CastFromHand): the entry came from the stack and the cast that put it
// there left the caster's hand. It is checked as the trigger fires and again
// as it resolves, so a reanimated, blinked or commander-zone Angel enters
// quietly. The sweep is every creature attacking as the trigger resolves,
// whoever controls it, exiled as one simultaneous event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "8b1e6b7c-892d-4058-b183-9121cdd623f0",
		Name:     "Angel of the Dire Hour",
		// ADR 0126 §6: only attacking creatures, and only when cast from your hand.
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && b30CastFromHand(g, source.InstanceID)
			}, "Angel of the Dire Hour — exile all attacking creatures", func(g *game.Game, item *game.StackItem) error {
				if !b30CastFromHand(g, item.SourceCardID) {
					return nil
				}
				return ExileAllMatching{Match: AttackingCreature()}.Apply(NewContext(g, item))
			}),
		},
	})
}
