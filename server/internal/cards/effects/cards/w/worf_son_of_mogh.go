package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worf, Son of Mogh — Legendary Creature — Klingon Warrior Officer
// {1}{R}{W}, 3/3 (Star Trek Commander, #4974):
//
//	"Menace
//	 Other Officers you control get +1/+1.
//	 Seek out — Whenever you draw your second card each turn, create a
//	 1/1 red Officer creature token."
//
// The anthem is Stromkirk Captain's tribal lord over Officers, read
// through the layered view (a changeling counts). The second-card
// trigger fires once a turn (trekB41SeekOut) and makes the printed
// Officer token, which the anthem then pumps.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2249af3-2e3e-484b-a31a-26432edd13ae",
		Name:            "Worf, Son of Mogh",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Officer"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.TrekB41SeekOut("Worf, Son of Mogh — create a 1/1 red Officer creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Officer"), N: 1})),
		},
	})
}
