package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tajuru Beastmaster — Creature — Elf Warrior Ally {5}{G}, 5/5 (Battle
// for Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control get +1/+1 until end of turn."
//
// Rally is one trigger per Ally entering under your control, this
// creature included (bfzB1Rally, Appa, the Vigilant's condition). The
// creatures it affects are fixed as it resolves (CR 611.2c), so a
// creature that enters later in the turn gets nothing from this
// instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc72ec48-9e8b-4ded-b445-f3a839fd1abe",
		Name:         "Tajuru Beastmaster",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BfzB1RallyTeamGets("Tajuru Beastmaster", 1)},
	})
}
