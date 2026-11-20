package effects

// Call to the Netherworld — Sorcery {B} (Time Spiral, #5596):
//
//	"Return target black creature card from your graveyard to your
//	 hand.
//	 Madness {0} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// A one-mana Raise Dead for black creature cards, free through madness (CR
// 702.35a). The card's colour is read in the graveyard, at announce and at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a28ef7a5-4b47-4429-838b-a64a7c0701d0",
		Name:         "Call to the Netherworld",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target black creature card from your graveyard", YouOwn(), Creature(), OfColor("B")),
		Madness:      "{0}",
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
