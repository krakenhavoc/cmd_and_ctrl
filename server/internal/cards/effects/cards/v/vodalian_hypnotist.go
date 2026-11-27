package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Hypnotist (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Merfolk Wizard {1}{U}, 1/1:
//	"{2}{B}, {T}: Target player discards a card. Activate only as a sorcery."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "586d64c7-6aa9-42a0-9d8a-9dece52bc5d6",
		Name:         "Vodalian Hypnotist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}, {T}: Target player discards a card. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Targets:      effects.TargetPlayer("target player"),
			SorcerySpeed: true,
			Effect:       effects.TargetPlayerDiscards(1),
		}},
	})
}
