package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bilbo, Retired Burglar — Legendary Creature — Halfling Rogue
// {1}{U}{R}, 1/3:
//
//	"When Bilbo enters or leaves the battlefield, the Ring tempts you.
//	 Whenever Bilbo deals combat damage to a player, create a Treasure
//	 token."
//
// One trigger for both events, as Cryogen Relic's. The leaves half
// looks back (CR 603.10a): its controller is whoever controlled Bilbo
// as it left, and Bilbo is gone by the time the tempt asks for a
// creature, so it can't be chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e7e3be5-04dd-4ff4-b121-3e65b3e00699",
		Name:         "Bilbo, Retired Burglar",
		Purpose:      game.Purpose{RingTempts: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrLeaves("Bilbo, Retired Burglar — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
			effects.WheneverThisDealsCombatDamageToAPlayer("Bilbo, Retired Burglar — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
