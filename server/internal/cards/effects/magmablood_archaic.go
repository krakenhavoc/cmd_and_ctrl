package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmablood Archaic — Creature — Avatar {2/R}{2/R}{2/R}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Reach, trample
//	 Converge — This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.
//	 Whenever you cast an instant or sorcery spell, creatures you control
//	 get +1/+0 until end of turn for each color of mana spent to cast that
//	 spell."
//
// Converge is an ability word. The entry counters are a CR 614.1c clause
// over the cast (Kurbis's EntersWithCountersFromCast) counting the distinct
// colours of mana spent (game.CastCounts.ColorsSpent); WantsDistinctColors
// pays each hybrid {2/R} with two generic mana of colours not yet spent when
// it can, and ReadsOwnPayment makes a waived payment ask what was paid.
//
// The cast trigger reads the colours of mana spent on the spell as it is cast
// (Aberrant Manawurm's Build) and carries the count on the item, so a spell
// countered in response still pumps; a spell cast for no mana pumps nothing.
// The set of creatures you control is locked as the trigger resolves (CR
// 611.2c), so a creature that enters afterwards gets nothing.
//
// No simplification.
func init() {
	const label = "Magmablood Archaic — creatures you control get +1/+0 until end of turn for each color of mana spent to cast that spell"
	t := WheneverYouCast(Or(Instant(), Sorcery()), label, magmabloodArchaicPump)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().ColorCount()
		return item
	}
	Register(Spec{
		OracleID:            "c6a5adc8-ee57-421e-8235-0b3160f44568",
		Name:                "Magmablood Archaic",
		Completeness:        CompletenessFull,
		PrintedKeywords:     []string{"reach", "trample"},
		ReadsManaSpent:      game.ReadsOwnPayment | game.ReadsYourCasts,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return cast.ColorsSpent },
		}},
		Triggered: []game.TriggeredAbility{t},
	})
}

// magmabloodArchaicPump gives each creature you control +1/+0 until end of
// turn for each colour of mana spent on the spell that triggered it.
func magmabloodArchaicPump(g *game.Game, item *game.StackItem) error {
	if item.Params.Amount <= 0 {
		return nil
	}
	return BoostUntilEOT{
		Match: And(Creature(), YouControl()),
		Power: item.Params.Amount,
		Label: "Magmablood Archaic — +1/+0 for each color of mana spent",
	}.Apply(NewContext(g, item))
}
