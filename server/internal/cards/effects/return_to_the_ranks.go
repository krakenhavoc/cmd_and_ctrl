package effects

// Return to the Ranks — Sorcery {X}{W}{W}:
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Return X target creature cards with mana value 2 or less from your
//	 graveyard to the battlefield."
//
// Convoke is the spell's tap cost (Chord of Calling), measured against the
// generic mana of X and the two {W}. The targets are exactly X creature
// cards in your graveyard (CountFromX, not "up to"), chosen as the spell is
// cast; every one still legal when it resolves comes back together as one
// entry (CR 608.2b, CR 603.6a), under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d4ee553-f206-4e70-a7b2-a3670440aee7",
		Name:         "Return to the Ranks",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets: targetsCountedByX(TargetCardInGraveyard("X target creature cards with mana value 2 or less from your graveyard",
			YouOwn(), Creature(), ManaValueLE(2))),
		OnResolve: cmdr2S33ReturnTargetedCardsTogether,
	})
}
