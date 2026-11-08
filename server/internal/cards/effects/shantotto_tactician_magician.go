package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shantotto, Tactician Magician — Legendary Creature — Dwarf Wizard
// {1}{U}{R}, 0/4 (Final Fantasy, #4462):
//
//	"Whenever you cast a noncreature spell, Shantotto gets +X/+0 until
//	 end of turn, where X is the amount of mana spent to cast that
//	 spell. If X is 4 or more, draw a card."
//
// X is read off the spell's payment record as it is cast and carried on
// the trigger (Aetherflux Conduit's Build), so a spell countered in
// response still counts. The pump lands only while Shantotto is still
// the object the trigger came from (CR 400.7); the draw does not depend
// on her, so it happens even if she has left.
//
// No simplification.
func init() {
	const label = "Shantotto — gets +X/+0 until end of turn; if X is 4 or more, draw a card"
	t := On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return YouCast(Noncreature())(ev, source, lki, g)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		x := item.Params.Amount
		if x > 0 && sourceIsStillThisPermanent(g, item) {
			if err := (BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Shantotto — +X/+0"}).Apply(ctx); err != nil {
				return err
			}
		}
		if x >= 4 {
			return DrawCards{N: 1}.Apply(ctx)
		}
		return nil
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:       "381df197-5fb7-44ab-a31f-96c4993a0e53",
		Name:           "Shantotto, Tactician Magician",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered:      []game.TriggeredAbility{t},
	})
}
