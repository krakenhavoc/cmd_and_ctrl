package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abuelo, Ancestral Echo — Legendary Creature — Spirit {1}{W}{U}, 2/2
// (The Lost Caverns of Ixalan, #4470):
//
//	"Flying, ward {2}
//	 {1}{W}{U}: Exile another target creature or artifact you control.
//	 Return it to the battlefield under its owner's control at the
//	 beginning of the next end step."
//
// Flying is printed and ward {2} is the shared ward trigger (Aegis
// Sculptor's). The activated ability is Roon of the Hidden Realm's
// blink with Saheeli, the Sun's Brilliance's target, "another target
// creature or artifact you control" (Another, so Abuelo itself is not a
// legal pick): the card is exiled as the ability resolves and a delayed
// trigger returns it under its owner's control at the beginning of the
// next end step (exileTargetsThenScheduleReturn). The returned card is a
// new object (CR 400.7). The ability has no tap in its cost, so it may be
// activated more than once a turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eaa9baac-72f2-4bde-9146-860878d9860b",
		Name:            "Abuelo, Ancestral Echo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Abuelo, Ancestral Echo — ward {2}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{U}: Exile another target creature or artifact you control. Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    effects.ManaCost("{1}{W}{U}"),
			Targets: effects.Another(effects.TargetPermanent("another target creature or artifact you control", effects.Or(effects.Creature(), effects.Artifact()), effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item),
					"Abuelo, Ancestral Echo — return the exiled card")
			},
		}},
	})
}
