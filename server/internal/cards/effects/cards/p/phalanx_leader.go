package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phalanx Leader — Creature — Human Soldier {W}{W}, 1/1 (Theros, S82 OLD b388):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put a +1/+1
//	 counter on each creature you control."
//
// Heroic is the cast trigger in shield_families_recipient.go (CR 207.2c). The set
// is the creatures the caster controls as the trigger resolves, counted before
// the first counter goes on (Basri's Solidarity's body), the Leader included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "857db7e0-254a-4920-a71d-9bab6555342f",
		Name:         "Phalanx Leader",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Phalanx Leader — put a +1/+1 counter on each creature you control (heroic)",
				effects.B11PutCounterOnEachCreatureYouControl),
		},
	})
}
