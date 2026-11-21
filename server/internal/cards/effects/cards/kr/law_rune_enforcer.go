package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Law-Rune Enforcer — Creature — Human Soldier {W}, 1/2 (Ravnica: Clue
// Edition, Ravnica Allegiance):
//
//	"{1}, {T}: Tap target creature with mana value 2 or greater."
//
// The {T} keeps it off the turn it arrives (CR 602.5a). The mana value is
// read as the ability is announced and again as it resolves (CR 608.2b), so
// a creature that stops qualifying in response is not tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd13392e-e72a-4dbb-ba9d-68f68e47e30e",
		Name:         "Law-Rune Enforcer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target creature with mana value 2 or greater.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with mana value 2 or greater", effects.ManaValueGE(2)),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
