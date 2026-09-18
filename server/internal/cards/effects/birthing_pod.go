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
// Both Phyrexian symbols are payable with 2 life now. #787 put the
// announce on the cast (CastSpellParams.PhyrexianLife, CR 107.4c) and
// #917 gave the ACTIVATION the same one
// (ActivateAbilityParams.PhyrexianLife, CR 602.2b) through the same
// strike-and-pay helper, so "{1}{G/P}" really is {1} and two life for
// a player without green. Nothing about it lives in this file, which
// is the point: a Phyrexian symbol is the cost engine's business, not
// a card's.
//
// Declared simplification: the board has no button for the CAST
// cost's {G/P} yet, so clicking this from hand still pays the {G}.
func init() {
	Register(Spec{
		OracleID:     "f8b9dd54-0837-47f4-ad14-7a0322d46d5f",
		Name:         "Birthing Pod",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The board has no button for the {G/P} in the casting cost yet — casting it from hand pays {G}, not 2 life. The activated ability's {G/P} can be paid either way."},
		Activated: []ActivatedAbility{{
			Label:        "{1}{G/P}, {T}, Sacrifice a creature: Search your library for a creature card with mana value equal to 1 plus the sacrificed creature's mana value, put it onto the battlefield, then shuffle.",
			Cost:         Plus(ManaCost("{1}{G/P}"), TapCost(), SacrificeACreature()),
			SorcerySpeed: true,
			Effect:       b19BirthingPodSearch,
		}},
	})
}
