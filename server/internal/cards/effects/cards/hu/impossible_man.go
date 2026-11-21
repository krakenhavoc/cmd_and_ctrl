package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "9be1d4b1-064d-4051-aafc-ca48def28e3b",
		Name:            "Impossible Man",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Impossible Man becomes a copy of another target permanent until end of turn, except his name is Impossible Man.",
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.Another(effects.TargetPermanent("another target permanent")),
			Effect: effects.SelfBecomesCopyOfTargetExcept("Impossible Man — becomes a copy until end of turn", func(v *game.PrintedValues) {
				v.SetName("Impossible Man")
			}),
		}},
	})
}
