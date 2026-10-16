package effects

// Surgical Extraction — Instant {B/P}:
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 Choose target card in a graveyard other than a basic land card.
//	 Search its owner's graveyard, hand, and library for any number of
//	 cards with the same name as that card and exile them. Then that
//	 player shuffles."
//
// The target is chosen, not exiled: it is found by the graveyard half of
// the search like any other card with its name, so the caster may leave
// it, and leave any other match too ("any number"). The caster searches
// its owner's zones in one prompt, seeing the whole hand as they do, and
// the owner's library is shuffled whatever was taken (extractionSearch,
// #4033). A target that has left the graveyard makes the spell do nothing
// (CR 608.2b).
//
// The Phyrexian symbol is the engine's: a cast may pay 2 life for it
// instead of {B} (CR 107.4f, #787).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3bfa088c-9a47-4123-aee5-4de44e355bef",
		Name:         "Surgical Extraction",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target card in a graveyard other than a basic land card",
			notABasicLandCard()),
		OnResolve: extractFromGraveyardTarget("Surgical Extraction", false),
	})
}
