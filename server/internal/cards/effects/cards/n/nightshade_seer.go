package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightshade Seer — Creature — Human Wizard {3}{B}, 1/1 (Urza's Destiny, S82
// #5621):
//
//	"{2}{B}, {T}: Reveal any number of black cards in your hand. Target creature
//	 gets -X/-X until end of turn, where X is the number of cards revealed this
//	 way."
//
// Scent of Nightshade's body as an activated ability (OldB706RevealAnyNumber): the
// target is chosen as the ability is activated, the reveal as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bcf66d0-8a4d-4c57-b4de-4cc35399465e",
		Name:         "Nightshade Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}: Reveal any number of black cards in your hand. Target creature gets -X/-X until end of turn, where X is the number of cards revealed this way.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetHarm}),
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB706AsEffect(effects.OldB706RevealAnyNumber("Nightshade Seer", "B", "choose-cards/s82-old-b706-nightshade-seer", effects.OldB706PumpByRevealed("Nightshade Seer", -1))),
		}},
	})
}
