package effects

// Agonasaur Rex — Creature — Dinosaur {3}{G}{G}, 8/8 (EDHREC rank
// 4528):
//
//	"Trample
//	 Cycling {2}{G} ({2}{G}, Discard this card: Draw a card.)
//	 When you cycle this card, put two +1/+1 counters on up to one
//	 target creature or Vehicle. It gains trample and indestructible
//	 until end of turn."
//
// An 8/8 trampler for five that is never a dead card: cycle it early
// for a card and a combat trick, cast it late as a threat. The
// flexibility is why it shows up in Dinosaur and stompy lists that
// would not otherwise play a five-mana vanilla body.
//
// DECLARED SIMPLIFICATION (weaker than printed): CYCLING is not
// implemented, and neither is the trigger that depends on it. Cycling
// is an activated ability that works only from HAND — "{2}{G},
// Discard this card: Draw a card" (CR 702.29a) — and the engine has
// neither a discard cost component nor activation from hand (#660),
// so there is no way to pay it. "When you cycle this card" triggers
// off that activation (CR 702.29c), so with no cycling there is
// nothing for it to fire on; declaring the trigger alone would
// attach it to nothing.
//
// The Rex therefore ships as an 8/8 trampler for five that has to be
// cast. That is strictly a subset of the printed card — a player
// loses the option, never gains one.
//
// The body half is complete: trample rides PrintedKeywords and feeds
// the combat engine's overflow math and the CR 510.1c damage-
// assignment prompt.
func init() {
	Register(Spec{
		OracleID:     "1ad5766b-9ae3-437b-bd91-c4c988a8b095",
		Name:         "Agonasaur Rex",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"It can't be cycled, so the cycling draw and the +1/+1 counters that come with it never happen — the Rex has to be cast.",
		},
		PrintedKeywords: []string{"trample"},
	})
}
