package effects

// Argivian Archaeologist — Creature — Human Artificer {1}{W}{W}, 1/1
// (Antiquities, #5640):
//
//	"{W}{W}, {T}: Return target artifact card from your graveyard to
//	 your hand."
//
// Buried Ruin's recursion on a creature, with a mana and a tap cost in
// place of the land's sacrifice: the same graveyard clause and the same
// resolution body (returnTargetedCardToHand). The tap is a {T} cost, so
// the Archaeologist cannot use it the turn it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b889ec0-6130-4e31-bb02-03fdabd28bee",
		Name:         "Argivian Archaeologist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}{W}, {T}: Return target artifact card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{W}{W}"), TapCost()),
			Targets: TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
			Effect:  returnTargetedCardToHand,
		}},
	})
}
