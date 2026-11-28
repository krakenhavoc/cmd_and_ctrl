package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ceremonial Guard — Creature — Human Soldier {2}{R}, 3/4 (Mercadian Masques, S82 OLD batch 687, #5619):
//
//	"When this creature attacks or blocks, destroy it at end of combat."
//
// Runaway Carriage's attack-or-block trigger (one trigger, however many attackers it blocks)
// over Cinder Wall's delayed destroy at the beginning of the end of combat step (CR 511.2),
// skipped if the Guard has left or come back as a new object (CR 400.7); regeneration or
// indestructible saves it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d8aae86-266f-4bfd-a058-0fe12b5ea354",
		Name:         "Ceremonial Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
				"Ceremonial Guard — destroy it at end of combat",
				effects.S82OldB364DestroyThisAtEndOfCombat("Ceremonial Guard — destroy it")),
		},
	})
}
