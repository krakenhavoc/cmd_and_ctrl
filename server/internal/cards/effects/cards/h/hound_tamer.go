package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hound Tamer // Untamed Pup — {2}{G} Creature — Human Werewolf 3/3 //
// Creature — Werewolf 4/4 (#2586, ADR 0132):
//
//	Front: "Trample
//	        {3}{G}: Put a +1/+1 counter on target creature.
//	        Daybound"
//	Back:  "Trample
//	        Other Wolves and Werewolves you control have trample.
//	        {3}{G}: Put a +1/+1 counter on target creature.
//	        Nightbound"
//
// The back face's grant is Village Reavers' lord shape with the "other"
// the card prints.
//
// No simplification.
func init() {
	const oracle = "e9208fc2-616d-4c32-bd66-76a8bf85a6b5"
	counter := func() effects.ActivatedAbility {
		return effects.ActivatedAbility{
			Label:   "{3}{G}: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PutPlusOneCounterOnEachLegalTarget,
		}
	}
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Hound Tamer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "daybound"},
		Activated:       []effects.ActivatedAbility{counter()},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Untamed Pup",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "nightbound"},
		Activated:       []effects.ActivatedAbility{counter()},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, Others: true, YoursOnly: true}, "trample"),
		},
	})
}
