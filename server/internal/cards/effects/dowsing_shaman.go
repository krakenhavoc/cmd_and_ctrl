package effects

// Dowsing Shaman — Creature — Centaur Shaman {4}{G}, 3/4 (Ravnica: City of
// Guilds, #5600):
//
//	"{2}{G}, {T}: Return target enchantment card from your graveyard to
//	 your hand."
//
// Argivian Archaeologist's row over an enchantment card: the target is
// checked again at resolution, so a card that left the graveyard in
// response stays where it went (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9cb34df-75b1-4dda-8f22-daadf59c5332",
		Name:         "Dowsing Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}, {T}: Return target enchantment card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{2}{G}"), TapCost()),
			Targets: TargetCardInGraveyard("target enchantment card in your graveyard", Enchantment(), YouOwn()),
			Effect:  returnTargetedCardToHand,
		}},
	})
}
