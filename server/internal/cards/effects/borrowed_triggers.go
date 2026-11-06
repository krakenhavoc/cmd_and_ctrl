package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// borrowed_triggers.go — the card side of ADR 0155's 2026-11-05
// amendment: "has all activated AND TRIGGERED abilities of …". The engine
// half is server/internal/game/borrowed_triggers.go; the "last chosen
// card" a lender can be is game/chosen_card.go.

// HasActivatedAndTriggeredAbilitiesOf is "~ has all activated and
// triggered abilities of …" (Koh, the Face Stealer): HasActivatedAbilitiesOf
// with AbilityBorrow.Triggered set, so the borrow also lends the lenders'
// triggered abilities, triggered mana abilities and triggered keywords.
func HasActivatedAndTriggeredAbilitiesOf(b game.AbilityBorrow) game.StaticAbility {
	b.Triggered = true
	s := HasActivatedAbilitiesOf(b)
	// The printed clause is the row's player-facing wording.
	s.Label = b.Label
	return s
}

// LastChosenExiledCardLender is "… of the last chosen card", where the
// card was chosen in exile by an ability of the borrowing source
// (Card.ChosenCard, Game.QueueChooseExiledCardForEffect): that card while it is
// still in exile as the object it was when chosen. Nothing chosen, or a
// chosen card that has left exile (a new object, CR 400.7), lends
// nothing. The static that uses it declares DependsOnExile, so the borrow
// is re-read the moment the card leaves exile; a new choice re-reads it
// too (SetChosenCardForEffect).
func LastChosenExiledCardLender() func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil {
			return nil
		}
		if c := g.ChosenExiledCardOf(source); c != nil && c.InstanceID != source.InstanceID {
			return []*game.Card{c}
		}
		return nil
	}
}
