package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dr. Pulaski, Cynical Clinician — Legendary Creature — Human Doctor
// {1}{W}, 2/2 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, you gain
//	 4 life."
//
// "Seek out" is an ability word (no rules meaning). The trigger is
// Gixian Puppeteer's: the second draw of each turn, any player's turn,
// and never the third (trekB31SeekOut).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "686d8449-e132-4478-b4f5-f5cadc8971c8",
		Name:         "Dr. Pulaski, Cynical Clinician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TrekB31SeekOut("Dr. Pulaski — gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
