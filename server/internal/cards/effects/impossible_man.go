package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impossible Man — Legendary Creature — Alien Shapeshifter {2}{U}, 1/4
// (Marvel Super Heroes Commander):
//
//	"Flying
//	 {2}{U}: Impossible Man becomes a copy of another target permanent until
//	 end of turn, except his name is Impossible Man."
//
// Mirage Mirror's duration copy (BecomeCopy, CR 707.2) with an "except"
// clause that sets the name. While he is the copy he has the copied
// permanent's characteristics and abilities and none of his own, so the
// ability on him is gone until cleanup puts him back (CR 611.2a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9be1d4b1-064d-4051-aafc-ca48def28e3b",
		Name:            "Impossible Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Impossible Man becomes a copy of another target permanent until end of turn, except his name is Impossible Man.",
			Cost:    ManaCost("{2}{U}"),
			Targets: Another(TargetPermanent("another target permanent")),
			Effect: selfBecomesCopyOfTargetExcept("Impossible Man — becomes a copy until end of turn", func(v *game.PrintedValues) {
				v.SetName("Impossible Man")
			}),
		}},
	})
}
