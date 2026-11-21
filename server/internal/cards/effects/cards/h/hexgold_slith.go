package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

var hexgoldSlithFirstStrikePaid = effects.EnergyPaid("pay-unless/hexgold-slith-first-strike",
	effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Hexgold Slith — first strike until end of turn"))

// Hexgold Slith — Creature — Slith {1}{W}, 2/1:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}{E}. If you do, it gains first strike until end of turn.
//	 Whenever this creature deals combat damage to a player, put a +1/+1 counter on it."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a467e75-f68d-4ad0-85ff-b8516de8d5bc",
		Name:         "Hexgold Slith",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Hexgold Slith", 2),
			effects.WhenThisAttacksMayPayEnergy("Hexgold Slith", 2, "have it gain first strike", hexgoldSlithFirstStrikePaid),
			effects.WheneverThisDealsCombatDamageToAPlayer("Hexgold Slith — put a +1/+1 counter on it", effects.ThisStillHere(effects.PlusOneCountersOnThis(1))),
		},
	})
}
