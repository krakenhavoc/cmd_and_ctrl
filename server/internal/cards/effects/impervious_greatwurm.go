package effects

// Impervious Greatwurm — 16/16 Creature — Wurm for {7}{G}{G}{G}
// (EDHREC rank 4419):
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of
//	 that creature's color.)
//	 Indestructible"
//
// Ten mana on paper and rather less in practice — a wide green board
// pays most of it with bodies, and what arrives is a 16/16 that
// wraths, Swords to Plowshares aside, simply do not answer. Roadmap
// batch 42 (#449), "no new machinery".
//
// Both halves are engine features rather than card code, which is
// what makes the file three fields long:
//
//   - Convoke (CR 702.51) is a cost COMPONENT, not an alternative
//     cost: the tapped creatures spend against the printed cost
//     instead of replacing it, each paying for {1} or for one mana of
//     its own colour. Spec.TapCost carries the pool of legal
//     permanents and that colour rule with it, which is the half a
//     hand-written cost would get wrong — and getting it wrong here
//     would ship the card stronger than printed.
//   - Indestructible (CR 702.12) is honoured by the destruction path
//     and by the two damage-driven creature state-based actions since
//     S25 (#77). It does not stop exile, sacrifice, -X/-X, or a
//     legend rule, and it never did.
//
// Convoke can pay coloured pips: three green creatures tapped to
// convoke cover {G}{G}{G}, which is what makes a token board able to
// cast this at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "65c0ab05-e740-4380-b39c-8bae9662f885",
		Name:            "Impervious Greatwurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		TapCost:         Convoke(),
	})
}
