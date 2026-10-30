package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wanda's Vision — {3}{R}{R} Enchantment (Marvel Super Heroes Commander,
// #4455):
//
//	"Whenever you cast your second spell each turn, exile cards from the
//	 top of your library until you exile a nonland card. You may cast that
//	 card without paying its mana cost."
//
// Ledger Shredder's "second spell" read for the enchantment's controller:
// the cast tally is bumped before the cast event goes out, so the second
// cast reads exactly two, and a spell cast by another player, or the
// first or third cast, does not trigger. The loop and the free cast are
// Nico Minoru's (exileUntilNonlandCastFree): made during the resolution
// (CR 608.2g), at flash timing, lapsing with the card still in exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1448eadb-e8b1-4e8f-9697-3bb919085dab",
		Name:         "Wanda's Vision",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && g.CastTallyFor(ev.Actor).Total == 2
			}, "Wanda's Vision — exile until a nonland card; you may cast it without paying its mana cost", wandasVisionExile),
		},
	})
}

func wandasVisionExile(g *game.Game, item *game.StackItem) error {
	return exileUntilNonlandCastFree(NewContext(g, item), "Wanda's Vision")
}
