package effects

// Birthing Pod — Artifact {3}{G/P} (EDHREC rank 2041):
//
//	"({G/P} can be paid with either {G} or 2 life.)
//	 {1}{G/P}, {T}, Sacrifice a creature: Search your library for a
//	 creature card with mana value equal to 1 plus the sacrificed
//	 creature's mana value, put that card onto the battlefield, then
//	 shuffle. Activate only as a sorcery."
//
// The chain-tutor. A CR 602 activation with three cost components
// and the sorcery-speed gate; the creature is paid at announce, and
// the body reads it back off the event log (b17PermanentSacrificedToPay,
// Jarad's read) to fix the one mana value the S22 search chooser
// accepts. A token's mana value is zero, so sacrificing one fetches
// a one-drop, as printed; the fetched creature enters untapped.
//
// Sandbox simplification, weaker than printed: the Phyrexian symbol
// in the ACTIVATION is charged as {G}. #787 put the "or 2 life" half
// on the cast (CastSpellParams.PhyrexianLife, CR 107.4c), so the
// spell's own {3}{G/P} can be paid with life by an engine caller;
// AbilityCost has no such announce, so the ability costs {1}{G},
// never {1} and two life. A cost-engine seam, not a card file's — and
// the board has no button for either half yet.
func init() {
	Register(Spec{
		OracleID:     "f8b9dd54-0837-47f4-ad14-7a0322d46d5f",
		Name:         "Birthing Pod",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The {G/P} in the activation must be paid with {G}, not with 2 life; an activated ability has no life-payment announce. The {G/P} in the cast cost is engine-payable with 2 life, but the board has no button for it."},
		Activated: []ActivatedAbility{{
			Label:        "{1}{G/P}, {T}, Sacrifice a creature: Search your library for a creature card with mana value equal to 1 plus the sacrificed creature's mana value, put it onto the battlefield, then shuffle.",
			Cost:         Plus(ManaCost("{1}{G/P}"), TapCost(), SacrificeACreature()),
			SorcerySpeed: true,
			Effect:       b19BirthingPodSearch,
		}},
	})
}
