package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infestation Expert // Infested Werewolf — {4}{G} Creature — Human
// Werewolf 3/4 // Creature — Werewolf 4/5 (#2561, ADR 0132):
//
//	Front: "Whenever this creature enters or attacks, create a 1/1 green
//	        Insect creature token.
//	        Daybound"
//	Back:  "Whenever this creature enters or attacks, create two 1/1
//	        green Insect creature tokens.
//	        Nightbound"
//
// The back face's "enters" matters: cast at night, the permanent enters
// on its back face (CR 702.145b), so it is Infested Werewolf's trigger,
// the two Insects, that sees it enter. A day-time transform never re-fires
// it (CR 712.18: nothing entered).
//
// No simplification.
func init() {
	const oracle = "bf69dc2a-9aec-4181-bbe9-70875055ec03"
	insects := func(name string, n int) game.TriggeredAbility {
		return effects.WhenThisEntersOrAttacks(name+" — create Insect tokens",
			effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Insect"), N: n}))
	}
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Infestation Expert",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered:       []game.TriggeredAbility{insects("Infestation Expert", 1)},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Infested Werewolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered:       []game.TriggeredAbility{insects("Infested Werewolf", 2)},
	})
}
