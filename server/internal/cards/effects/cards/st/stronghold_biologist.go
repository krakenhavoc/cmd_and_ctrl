package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stronghold Biologist — Creature — Human Spellshaper {2}{U}, 1/1
// (Nemesis, #5618):
//
//	"{U}{U}, {T}, Discard a card: Counter target creature spell."
//
// The spellshaper cost (nemSpellshaperCost) on Tidal Control's body,
// counterTheChosenSpell: the spell is countered if it is still a legal
// target as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e2c4508-54ad-45c2-b35f-678f8c376b27",
		Name:         "Stronghold Biologist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}, {T}, Discard a card: Counter target creature spell.",
			Cost:    effects.NemSpellshaperCost("{U}{U}"),
			Targets: effects.TargetSpell("target creature spell", effects.Creature()),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
