package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firemantle Mage — Creature — Human Shaman Ally {2}{R}, 2/2 (Battle
// for Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control gain menace until end of turn. (A
//	 creature with menace can't be blocked except by two or more
//	 creatures.)"
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
		OracleID:     "27599eff-19df-48e2-9a1e-0c087f42a075",
		Name:         "Firemantle Mage",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BfzB1RallyTeamGains("Firemantle Mage", "menace")},
	})
}
