package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// leftMulticolored reports whether the permanent an EventLTB names had two or
// more colours as it last existed on the battlefield (CR 603.10a), read
// through leftAsColor for each of the five.
func leftMulticolored(ev game.Event, c game.Card) bool {
	n := 0
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		if leftAsColor(ev, c, color) {
			n++
		}
	}
	return n >= 2
}

// nehebDiscardDrawMana is Neheb, Dreadhorde Champion's "if you do, draw that
// many cards and add that much {R}": a keyed discard continuation (#4264) of
// the count really discarded. The cards are drawn first, then the red mana is
// added, marked to last until end of turn.
var nehebDiscardDrawMana = discardDraw{then: game.RegisterDiscardThenRule("choose-cards/discard-draw-and-add-red-kept", nil,
	func(g *game.Game, r game.DiscardsDone) error {
		k := r.Discarded.Count()
		if k <= 0 {
			return nil
		}
		if err := g.DrawNForEffect(r.Player, k); err != nil {
			return err
		}
		return g.AddManaWithOptionsForEffect(r.Player, r.Source, "", game.AddManaOptions{
			Amount: ManaN(k, "R"),
			Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
		})
	})}
