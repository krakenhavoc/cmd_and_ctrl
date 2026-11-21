package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tatyova, Benthic Druid — Legendary Creature — Merfolk Druid
// {3}{G}{U}, 3/3 (EDHREC rank 472):
//
//	"Landfall — Whenever a land you control enters, you gain 1 life
//	 and draw a card."
//
// The Simic lands commander: every land drop, fetch and ramp spell
// is a card. Landfall is the ETB trigger filtered to lands, firing
// for any way a land enters under your control (Tireless
// Provisioner's shape); a fetchland fires it twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0715e860-3b3b-4331-9718-207973e94fee",
		Name:         "Tatyova, Benthic Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Tatyova — gain 1 life and draw a card (landfall)", game.LandfallPayoff{Draws: 1, LifeGain: 1},
				effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1})),
		},
	})
}
