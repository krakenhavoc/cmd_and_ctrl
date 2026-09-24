package effects

// Herd Heirloom — Artifact {1}{G}:
//
//	"{T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell.
//	 {T}: Until end of turn, target creature you control with power 4
//	 or greater gains trample and 'Whenever this creature deals combat
//	 damage to a player, draw a card.'"
//
// The mana ability is Ancient Ziggurat's restricted-spend shape
// (ManaRestrictType("Creature")) with a full colour pipe instead of a
// fixed colour.
//
// Caveat: the second ability isn't implemented. It grants a NEW
// triggered ability to another permanent for a limited duration — the
// still-open half of "Abilities granted to other permanents"
// (docs/engine-seams.md, ADR 0093, #754): static and attached grants
// ship, but granted-TRIGGER cards and duration grants from a
// resolving ability are both listed as still missing. Only the mana
// ability works.
func init() {
	Register(Spec{
		OracleID:     "78b6dd40-c182-4037-a4d3-9fd012b2c584",
		Name:         "Herd Heirloom",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The {T}: Until end of turn, target creature you control with power 4 or greater gains trample and \"Whenever this creature deals combat damage to a player, draw a card.\" ability isn't implemented — granting a triggered ability from a resolving ability, for a limited duration, is still open engine machinery. Only the mana ability works."},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{ManaRestrictType("Creature")},
			Label:        "Add one mana of any color. Spend this mana only to cast a creature spell",
		}},
	})
}
