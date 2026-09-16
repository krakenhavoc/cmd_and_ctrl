package effects

// Phyrexian Reclamation — Enchantment, {B} (EDHREC rank 948):
//
//	"{1}{B}, Pay 2 life: Return target creature card from your
//	 graveyard to your hand."
//
// The one-mana recursion engine: two mana and two life a creature,
// as many times a turn as you can pay. One activated ability with a
// mana-and-life cost, validated together before either is paid, and
// a graveyard target clause answered by clicking the card in the
// zone browser — Buried Ruin's shape with a creature predicate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "647ca69e-cc01-4b2b-b376-bee2a98331e8",
		Name:         "Phyrexian Reclamation",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Pay 2 life: Return target creature card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{1}{B}"), PayLife(2)),
			Targets: TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn()),
			Effect:  returnTargetCardToHand,
		}},
	})
}
