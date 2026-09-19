package effects

// Secluded Courtyard — Land:
//
//	"As this land enters, choose a creature type.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell of the chosen type or activate an ability of a
//	 creature source of the chosen type."
//
// Cavern of Souls without the counter protection, and the card the
// batch-01 re-triage flagged with the refutation checklist for its
// "Spend this mana only … or". The "or" turns out to need nothing new:
// a restriction list is a set of tags, ANDed, and Cavern's list names
// a PURPOSE (casts only) where this one does not. Drop that one tag
// and the remaining pair — "the object is a Creature" and "the object
// has the named subtype" — is satisfied by a cast OR an activation,
// because CR 106.6a measures an activation against the ability's
// SOURCE permanent. That is the printed sentence exactly; see
// ChosenTypeCastOrActivateManaRestrictions in tribal.go.
//
// Two abilities, not one with a mode, for Cavern's reason: the
// colourless half is unrestricted, so the auto-tapper can plan around
// it, and the coloured half carries the restriction and is therefore
// invisible to the planner — which spell restricted mana is for is a
// decision the planner cannot make.
//
// "Any color" flatly: the printed text does not mention the
// commander's identity, so the pipe keeps all five and
// NarrowToCommanderIdentity stays off.
//
// Until the controller answers the entry prompt the named type is
// empty and the coloured mana is unspendable — the weaker direction,
// and the same behaviour Cavern has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79ba18fd-f184-43c1-86df-56ee18ce806c",
		Name:         "Secluded Courtyard",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Secluded Courtyard"),
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:             ManaAbilityCost{Tap: true},
				Produced:         "{W|U|B|R|G}",
				Label:            "Add one mana of any color (chosen-type creatures only)",
				RestrictionsFunc: ChosenTypeCastOrActivateManaRestrictions(),
			},
		},
	})
}
