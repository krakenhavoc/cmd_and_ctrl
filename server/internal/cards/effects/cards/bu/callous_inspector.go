package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Callous Inspector — Creature — Human Soldier {B}, 1/1:
//
//	"Menace
//	 When this creature dies, it deals 1 damage to you. Create a Clue
//	 token."
//
// One death trigger with two sentences. The damage comes from the
// Inspector as it last existed (CR 608.2h), so it is damage from a black
// creature, not life loss. The Clue follows whether or not the damage
// was prevented; creating it is not investigating (the ruling).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d950fd4-8956-4b1a-ab0f-ab31bcffce4e",
		Name:            "Callous Inspector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Callous Inspector — it deals 1 damage to you; create a Clue token", effects.Do(
				effects.DamageController{N: 1},
				effects.CreateToken{Template: effects.ClueToken(), N: 1},
			)),
		},
	})
}
