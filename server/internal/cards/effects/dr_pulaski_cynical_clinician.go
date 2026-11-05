package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "686d8449-e132-4478-b4f5-f5cadc8971c8",
		Name:         "Dr. Pulaski, Cynical Clinician",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			trekB31SeekOut("Dr. Pulaski — gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
