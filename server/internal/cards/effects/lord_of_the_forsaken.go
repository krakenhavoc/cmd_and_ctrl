package effects

// Lord of the Forsaken — Creature — Demon {4}{B}{B}, 6/6 (Tarkir: Dragonstorm Commander)
// Commander):
//
//	"Flying, trample
//	 {B}, Sacrifice another creature: Target player mills three cards.
//	 Pay 1 life: Add {C}. Spend this mana only to cast a spell from your
//	 graveyard."
//
// The mill is a plain activated ability whose sacrifice is a cost. The
// mana ability costs life, not a tap, and its mana carries the
// graveyard-cast tag (ManaRestrictCastFromGraveyard, the twin of Web
// Watch's exile tag): it pays for a cast whose zone is a graveyard and
// for nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a25d1e2-63d8-4a93-b0ce-f9307736b3f9",
		Name:            "Lord of the Forsaken",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Life: 1},
			Produced:     "{C}",
			Label:        "Pay 1 life: Add {C}. Spend this mana only to cast a spell from your graveyard",
			Restrictions: []string{ManaRestrictCastFromGraveyard},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice another creature: Target player mills three cards.",
			Cost:    Plus(ManaCost("{B}"), SacrificeAnotherN(1, "another creature", Creature())),
			Targets: TargetPlayer("target player"),
			Effect:  TargetPlayerMills(3),
		}},
	})
}
