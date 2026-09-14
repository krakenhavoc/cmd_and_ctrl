package effects

// Eden, Seat of the Sanctum — Land — Town (EDHREC rank 3714):
//
//	"{T}: Add {C}.
//	 {5}, {T}: Mill two cards. Then you may sacrifice this land. When
//	 you do, return another target permanent card from your graveyard
//	 to your hand."
//
// A colourless utility land that turns into a Regrowth for
// permanents. The mana is a plain {C}. The printed second ability is
// one activation with a choice in the middle of it and a reflexive
// trigger after the choice, and the engine has neither a yes/no
// prompt a resolving ability can ask its controller nor a way to put
// a targeted reflexive trigger on the stack from a body — so it is
// shipped as TWO CR 602 activations that share the {5} and the tap
// (Insidious Fungus's split), and the choice is made at activation:
//
//   - "{5}, {T}: Mill two cards." — the printed ability with the
//     sacrifice declined.
//   - "{5}, {T}, Sacrifice this land: Mill two cards, then return
//     another target permanent card from your graveyard to your
//     hand." — the printed ability with the sacrifice accepted.
//     Codex Shredder's second ability with a mill in front: the
//     target is a permanent card in the controller's graveyard,
//     picked at announce (CR 601.2c) and re-checked at resolution;
//     Eden is sacrificed at announce after the pick, so "another" is
//     automatic, exactly as in paper.
//
// Both simplifications run the weaker way: the controller decides
// whether to sacrifice before seeing the two milled cards rather
// than after, and the card to return is chosen before the mill, so
// the two cards that activation mills can never be it (printed, the
// reflexive trigger's target is chosen after the mill and may be one
// of them). Nothing is ever returned that the printed card could not
// return.
func init() {
	Register(Spec{
		OracleID:     "84856b92-5ce8-47f3-9a1c-78d6a3e26aca",
		Name:         "Eden, Seat of the Sanctum",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Sacrificing is a separate activation you choose up front, not a decision made after seeing the two milled cards.",
			"The permanent card to return is chosen when you activate, before the mill, so the two cards milled by that activation can't be picked.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:  "{5}, {T}: Mill two cards.",
				Cost:   Plus(ManaCost("{5}"), TapCost()),
				Effect: b35MillTwo,
			},
			{
				Label:   "{5}, {T}, Sacrifice this land: Mill two cards, then return another target permanent card from your graveyard to your hand.",
				Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
				Targets: TargetCardInGraveyard("another target permanent card from your graveyard", YouOwn(), Permanent()),
				Effect:  b35MillTwoThenReturnChosen,
			},
		},
	})
}
