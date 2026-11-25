package game

// mana_ability_x.go — #6125: X in a mana ability's cost (CR 107.3a; a
// mana ability is an activated ability, CR 605.1a, activated by CR
// 602.2b's process, so its controller announces X with the activation).
//
//   - Wizard's Rockets: "{X}, {T}, Sacrifice this artifact: Add X mana in
//     any combination of colors." The {X} is in the mana cost.
//   - Springjack Pasture: "{T}, Sacrifice X Goats: Add X mana of any one
//     color. You gain X life." The X counts a sacrifice.
//
// The activator announces X on ManaAbilityParams.XValue; the mana cost is
// priced and the sacrifice counted with it, and the ability resolves at
// once (CR 605.3b) reading it off PaidCost.XValue. The auto-tapper never
// plans such an ability: how much to pay is the player's decision.

// DemandsX reports whether this mana ability's cost has an X to announce:
// {X} in its mana cost, or a "Sacrifice X …" clause.
func (a ManaAbilityShape) DemandsX() bool {
	if SacrificeCountFromX(a.SacrificeOther) {
		return true
	}
	if a.ManaCost == "" {
		return false
	}
	cost, err := ParseCost(a.ManaCost)
	return err == nil && cost.XSlots > 0
}
