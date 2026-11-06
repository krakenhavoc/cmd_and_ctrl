package effects

// Refute — Instant {1}{U}{U} (EDHREC rank 1899):
//
//	"Counter target spell. Draw a card, then discard a card."
//
// Cancel with a loot stapled on. The counter is the ordinary
// CounterTarget; the loot is lootOne, which draws first and queues
// the discard second so the drawn card is a legal discard, as in
// paper. If the target left the stack in response, Refute is
// countered by game rules before it resolves (CR 608.2b) and the
// loot never happens — also as in paper.
//
// "Counter target SPELL": an activated or triggered ability on the
// stack is not a legal target of this card as printed, so the old
// caveat ("only spells can be countered") described the card, not a
// gap. TargetSpell is the printed clause.
func init() {
	Register(Spec{
		OracleID:     "eb1dfa29-7371-4cb6-bfa2-16f7820b69be",
		Name:         "Refute",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve:    counterTheTargetThenLootOne,
	})
}
