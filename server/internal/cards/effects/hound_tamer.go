package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	counter := func() ActivatedAbility {
		return ActivatedAbility{
			Label:   "{3}{G}: Put a +1/+1 counter on target creature.",
			Cost:    ManaCost("{3}{G}"),
			Targets: TargetCreature("target creature"),
			Effect:  putPlusOneCounterOnEachLegalTarget,
		}
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Hound Tamer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "daybound"},
		Activated:       []ActivatedAbility{counter()},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Untamed Pup",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "nightbound"},
		Activated:       []ActivatedAbility{counter()},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, Others: true, YoursOnly: true}, "trample"),
		},
	})
}
