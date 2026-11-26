package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kalitas, Bloodchief of Ghet — Legendary Creature — Vampire Warrior {5}{B}{B},
// 5/5 (Zendikar, S82 OLD b496):
//
//	"{B}{B}{B}, {T}: Destroy target creature. If that creature dies this way,
//	 create a black Vampire creature token. Its power is equal to that creature's
//	 power and its toughness is equal to that creature's toughness."
//
// Saw in Half's destroy-then-token: the creature's power and toughness are read
// as it last existed on the battlefield, before the destroy; "dies this way" is
// read off the leaves-the-battlefield event the destruction produced, so a
// regenerated, indestructible or exiled-instead creature makes no token, and a
// token creature that was destroyed does (it dies, CR 700.4). The Vampire is the
// controller's, an X/X whose printed size is */* (the 0/0 template is stamped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "003bf21f-6815-4098-a186-aa15f5947541",
		Name:         "Kalitas, Bloodchief of Ghet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}{B}, {T}: Destroy target creature. If that creature dies this way, create a black Vampire creature token. Its power is equal to that creature's power and its toughness is equal to that creature's toughness.",
			Cost:    effects.Plus(effects.ManaCost("{B}{B}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.Old496KalitasDestroy,
		}},
	})
}
