package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kessig Naturalist // Lord of the Ulvenwald — {R}{G} Creature — Human
// Werewolf 2/2 // Creature — Werewolf 3/3 (#2586, ADR 0132):
//
//	Front: "Whenever this creature attacks, add {R} or {G}. Until end of
//	        turn, you don't lose this mana as steps and phases end.
//	        Daybound"
//	Back:  "Other Wolves and Werewolves you control get +1/+1.
//	        Whenever this creature attacks, add {R} or {G}. Until end of
//	        turn, you don't lose this mana as steps and phases end.
//	        Nightbound"
//
// The mana is Savage Ventmaw's keep-until-end-of-turn mana with a pipe
// pick for the colour. The back face's lord is the Goblin King shape,
// "other" and "you control" both printed.
//
// No simplification.
func init() {
	const oracle = "54f8acb1-58f3-49d1-bff8-c1b578245936"
	mana := func(name string) game.TriggeredAbility {
		return effects.WheneverThisAttacks(name+" — add {R} or {G}", effects.Do(effects.AddMana{
			Produced: "{R|G}",
			Riders:   []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()},
		}))
	}
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Kessig Naturalist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered:       []game.TriggeredAbility{mana("Kessig Naturalist")},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Lord of the Ulvenwald",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered:       []game.TriggeredAbility{mana("Lord of the Ulvenwald")},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
