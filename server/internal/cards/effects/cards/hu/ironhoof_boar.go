package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironhoof Boar — Artifact Creature — Boar {5}{R}, 5/4 (Neon Dynasty, #5493):
//
//	"Trample, haste
//	 Channel — {1}{R}, Discard this card: Target creature gets +3/+1 and
//	 gains trample until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "90266ce6-4360-4de4-9b06-adff7d6c378b",
		Name:            "Ironhoof Boar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label: "Channel — {1}{R}, Discard this card: Target creature gets +3/+1 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(game.TargetPurpose{
				Slot: 0, UntilEndOfTurn: game.TargetBenefit,
				Pump: &game.Pump{Power: 3, Toughness: 1, Keywords: []string{"trample"}},
			})},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Ironhoof Boar", 3, 1, "trample"),
		}},
	})
}
