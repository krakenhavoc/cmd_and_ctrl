package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disrupting Scepter — Artifact {3} (Eighth Edition, S82 OLD batch 612, #5607):
//
//	"{3}, {T}: Target player discards a card. Activate only during your turn."
//
// Scepter of Fugue's ability at {3}: the DuringYourTurn activation condition
// (CR 602.5b) lets it be activated at instant speed in any step of your own
// turn, never on an opponent's; the discarding player chooses the card
// (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd30a128-4b23-476e-8066-2272fdc395d9",
		Name:         "Disrupting Scepter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}, {T}: Target player discards a card. Activate only during your turn.",
			Purpose:   effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:      effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets:   effects.TargetPlayer("target player"),
			Condition: effects.DuringYourTurn(),
			Effect:    effects.TargetPlayerDiscards(1),
		}},
	})
}
