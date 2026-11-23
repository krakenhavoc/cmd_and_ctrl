package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimblewright Schematic — Artifact {2} (Commander Legends: Battle for
// Baldur's Gate, #5490):
//
//	"When this artifact enters or is put into a graveyard from the
//	 battlefield, create a 1/1 colorless Construct artifact creature
//	 token."
//
// One printed ability with two trigger conditions, so one row watching
// both kinds (Mogg War Marshal's WhenThisEntersOrDies). "Is put into a
// graveyard from the battlefield" is what dies means (CR 700.4), and
// the leave half is gated on the graveyard, so a bounce or an exile does
// not make a Construct. The token is the table's 1/1 colorless Construct
// artifact creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b0dcf09-df51-4ad3-848e-c0ffb25824f9",
		Name:         "Nimblewright Schematic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Nimblewright Schematic — create a 1/1 colorless Construct artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Construct artifact"), N: 1})),
		},
	})
}
