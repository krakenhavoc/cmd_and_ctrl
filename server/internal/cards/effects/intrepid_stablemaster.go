package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intrepid Stablemaster — Creature — Human Scout {1}{G}, 2/2 (Outlaws of Thunder Junction):
//
//	"Reach
//	 {T}: Add {G}.
//	 {T}: Add two mana of any one color. Spend this mana only to cast Mount
//	 or Vehicle spells."
//
// The second ability is Alchemist's Talent's two mana of one colour with a
// spend restriction (CR 106.6) to casting a spell with the subtype Mount or
// Vehicle; any other use is refused. The first is ordinary.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a825bdea-3ea2-46b2-93b9-21586643aab0",
		Name:            "Intrepid Stablemaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "{T}: Add {G}.",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     OneColorOfAmount(2),
				Label:        "{T}: Add two mana of any one color. Spend this mana only to cast Mount or Vehicle spells.",
				Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Mount|Vehicle")},
			},
		},
	})
}
