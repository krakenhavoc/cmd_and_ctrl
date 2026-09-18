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
// The two keywords are printed strings. The X counters take
// Benevolent Hydra's posture, and it carries the same declared
// simplification for the same engine reason.
//
// DECLARED SIMPLIFICATION, weaker than printed: the counters are
// placed as the SPELL RESOLVES, a beat before the creature enters,
// because an entry replacement cannot read the spell's X. The
// creature is on the battlefield at the right size before anything
// can act — a 0/0 never faces a state-based check — so the difference
// is invisible except to a payoff that watches counters being PUT on
// a permanent (Corpsejack Menace doubles them as printed, but a
// "whenever one or more counters are put on a creature you control"
// trigger sees them against a card that is technically still
// resolving). Nothing gains from the gap.
func init() {
	Register(Spec{
		OracleID:        "b7daa74c-6142-4107-9355-be98af6ccf13",
		Name:            "Shivan Devastator",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The X +1/+1 counters are put on the Dragon as the spell resolves, a moment before it enters, so effects that watch you put counters on a permanent may not see them."},
		PrintedKeywords: []string{"flying", "haste"},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
		},
	})
}
