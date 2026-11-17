package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dromad Purebred — Creature — Camel Beast {4}{W}, 1/5 (Ravnica: City of
// Guilds, #5600):
//
//	"Whenever this creature is dealt damage, you gain 1 life."
//
// Boros Reckoner's damage trigger: once for each batch of damage dealt at
// the same time, whatever the amount (CR 120.3, CR 510.2), and never for 0
// damage (CR 120.8). Its controller gains the life even if the damage
// killed it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7adac8f-8187-49a0-80b2-3281d42f1b92",
		Name:         "Dromad Purebred",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Dromad Purebred — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
