package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ekthi, Contaminator Priest — Legendary Creature — Phyrexian Cleric {3}{W},
// 3/3 (Mystery Booster Commander, #5948):
//
//	"Other Phyrexians you control get +1/+1.
//	 Each Equipment you control has living weapon. (Whenever an Equipment you
//	 control enters, create a 0/0 black Phyrexian Germ creature token, then
//	 attach that Equipment to it.)"
//
// The anthem is Anthem of Champions' layer-7c static over the other Phyrexian
// permanents you control; the Germ the living weapon makes is itself a
// Phyrexian, so it is a 1/1 while Ekthi is out. The second sentence is an
// ADR 0093 grant of the shared LivingWeapon trigger to each Equipment you
// control (Candlekeep Sage's shape): the Equipment has the ability as it
// enters while Ekthi is on the battlefield, so its entering triggers it
// (CR 603.6a), and one that already has living weapon makes two Germs, one
// for each instance. An Equipment that entered before Ekthi does nothing.
//
// No simplification.
const ekthiLivingWeaponGrant = "ekthi-contaminator-priest/living-weapon"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "e298a2a4-0e04-4af9-8638-7c01668d97d0",
		Name:         "Ekthi, Contaminator Priest",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       ekthiLivingWeaponGrant,
			Triggered: []game.TriggeredAbility{effects.LivingWeapon("This Equipment")},
			Text:      "Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Phyrexian"))
			}, 1, 1), "Other Phyrexians you control get +1/+1."),
			effects.Labelled(effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype("Equipment")
			}, ekthiLivingWeaponGrant), "Each Equipment you control has living weapon."),
		},
	})
}
