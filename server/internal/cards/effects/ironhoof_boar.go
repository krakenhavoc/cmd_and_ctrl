package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ironhoof Boar — Artifact Creature — Boar {5}{R}, 5/4 (Neon Dynasty, #5493):
//
//	"Trample, haste
//	 Channel — {1}{R}, Discard this card: Target creature gets +3/+1 and
//	 gains trample until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "90266ce6-4360-4de4-9b06-adff7d6c378b",
		Name:            "Ironhoof Boar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []ActivatedAbility{{
			Label: "Channel — {1}{R}, Discard this card: Target creature gets +3/+1 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(game.TargetPurpose{
				Slot: 0, UntilEndOfTurn: game.TargetBenefit,
				Pump: &game.Pump{Power: 3, Toughness: 1, Keywords: []string{"trample"}},
			})},
			Cost:    Plus(ManaCost("{1}{R}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature"),
			Effect:  neoB07TargetGetsUntilEOT("Ironhoof Boar", 3, 1, "trample"),
		}},
	})
}
