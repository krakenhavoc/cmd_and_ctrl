package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivy Seer — Creature — Elf Wizard {3}{G}, 1/1 (Urza's Destiny, S82 #5621):
//
//	"{2}{G}, {T}: Reveal any number of green cards in your hand. Target creature
//	 gets +X/+X until end of turn, where X is the number of cards revealed this
//	 way."
//
// Scent of Ivy's body as an activated ability (OldB706RevealAnyNumber): the
// target is chosen as the ability is activated, the reveal as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f61591d5-0f2a-48cd-a091-863d2e93e019",
		Name:         "Ivy Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}: Reveal any number of green cards in your hand. Target creature gets +X/+X until end of turn, where X is the number of cards revealed this way.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB706AsEffect(effects.OldB706RevealAnyNumber("Ivy Seer", "G", "choose-cards/s82-old-b706-ivy-seer", effects.OldB706PumpByRevealed("Ivy Seer", 1))),
		}},
	})
}
