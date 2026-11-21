package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jorubai Murk Lurker — Creature — Leech {2}{U}, 1/3 (Magic 2015, #5556):
//
//	"This creature gets +1/+1 as long as you control a Swamp.
//	 {1}{B}: Target creature gains lifelink until end of turn."
//
// Kird Ape's land-gated pump with a Swamp, and Alabaster Mage's lifelink
// grant on any creature (jmpB7FirstLegalTargetGains), the target
// re-checked at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12e091d0-3fd0-4058-84a9-529023b74c74",
		Name:         "Jorubai Murk Lurker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Swamp.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Swamp")),
			Apply:     effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Target creature gains lifelink until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    effects.ManaCost("{1}{B}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.JmpB7FirstLegalTargetGains("Jorubai Murk Lurker — lifelink until end of turn", "lifelink"),
		}},
	})
}
