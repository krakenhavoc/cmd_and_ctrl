package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brine Seer — Creature — Human Wizard {3}{U}, 1/1 (Urza's Destiny, S82 #5621):
//
//	"{2}{U}, {T}: Reveal any number of blue cards in your hand. Counter target
//	 spell unless its controller pays {1} for each card revealed this way."
//
// Scent of Brine's body as an activated ability (OldB706RevealAnyNumber): the
// target spell is chosen as the ability is activated, the reveal as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64c88c94-3a9d-4fd4-b58c-4d2167c2b8d5",
		Name:         "Brine Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Reveal any number of blue cards in your hand. Counter target spell unless its controller pays {1} for each card revealed this way.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.OldB706AsEffect(effects.OldB706RevealAnyNumber("Brine Seer", "U", "choose-cards/s82-old-b706-brine-seer", effects.OldB706CounterUnlessPaysPerRevealed("Brine Seer"))),
		}},
	})
}
