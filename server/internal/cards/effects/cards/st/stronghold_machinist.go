package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stronghold Machinist — Creature — Human Spellshaper {2}{U}, 1/1
// (Nemesis, #5618):
//
//	"{U}{U}, {T}, Discard a card: Counter target noncreature spell."
//
// The spellshaper cost (nemSpellshaperCost) on Tidal Control's body,
// counterTheChosenSpell: the spell is countered if it is still a legal
// target as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50bbbb39-4709-478f-bb51-716c7e4f8323",
		Name:         "Stronghold Machinist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}, {T}, Discard a card: Counter target noncreature spell.",
			Cost:    effects.NemSpellshaperCost("{U}{U}"),
			Targets: effects.TargetSpell("target noncreature spell", effects.Noncreature()),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
