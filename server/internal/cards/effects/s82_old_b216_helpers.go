package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b216_helpers.go — shared pieces for Rivals of Ixalan batch 216
// (S82, #5531): the Forerunner cycle's tribal enters triggers.

// S82OldB216TribeEnters is "whenever [another] <subtype> you control enters"
// for a trigger's AppliesTo: the entering permanent is controlled by the
// trigger's controller, has the subtype as it enters (a changeling counting
// as every type, CR 702.73a), and is not the source itself when `another`.
func S82OldB216TribeEnters(subtype string, another bool) func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := EnteredUnderYourControl(ev, source, g, another)
		return ok && c.HasSubtypeWord(subtype)
	}
}

// S82OldB216SearchForAllNamed is "you may search your library for any number
// of cards named <name>, reveal them, put them into your hand, then shuffle"
// (Legion Conquistador; Battalion Foot Soldier prints the same sentence): the
// "you may" and the count are the search's own, zero being a legal answer.
func S82OldB216SearchForAllNamed(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: func(c game.Card) bool { return c.Name == name },
			Dest:      game.ZoneHand,
			Unbounded: true,
			Reveal:    true,
			Shuffle:   true,
			Optional:  true,
			Reason:    name + " — you may search for any number of cards named " + name,
		}.Apply(NewContext(g, item))
	}
}
