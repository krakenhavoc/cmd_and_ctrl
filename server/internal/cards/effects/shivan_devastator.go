package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Devastator — Creature — Dragon Hydra {X}{R}, 0/0 (EDHREC
// rank 4164):
//
//	"Flying, haste
//	 This creature enters with X +1/+1 counters on it."
//
// A red X-spell that is also a Dragon and a Hydra, which is why it is
// played in three different decks at once: it is a finisher that
// attacks the turn it lands, a tribal body for Miirym or Gargos
// (also in this batch), and a counter target for a Doubling Season.
// Haste is what separates it from every other Hydra — X mana of
// damage the turn you cast it, out of nowhere.
//
// The two keywords are printed strings; the X counters are the
// printed entry clause below.
//
// The X counters are the printed CR 614.1c entry clause and ride the
// CR 614 pipeline as one — XCounters, seeded onto the entry event off
// the resolving stack item while the spell is still there (#1002).
// They land on the PERMANENT, after the move and before EventETB, so
// Doubling Season and Hardened Scales apply, the card's own enters
// trigger reads a finished creature, and a "whenever one or more
// counters are put on a permanent you control" payoff sees them —
// which it could not while they went onto a card still on the stack.
//
// Cast for X=0 the Dragon enters as the printed 0/0 it is and the
// next state-based check puts it into its owner's graveyard
// (CR 704.5f), as in paper — an engine gap until #691, when the
// printing behind the object became what the toughness check reads
// (game.Card.ToughnessIsKnown).
func init() {
	Register(Spec{
		OracleID:                   "b7daa74c-6142-4107-9355-be98af6ccf13",
		Name:                       "Shivan Devastator",
		XMatters:                   true,
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying", "haste"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
	})
}
