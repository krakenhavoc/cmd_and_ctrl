package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canker Abomination — {2}{B/G}{B/G} Creature — Treefolk Horror, 6/6:
//
//	"As this creature enters, choose an opponent. This creature enters with a
//	 -1/-1 counter on it for each creature that player controls."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4440e41c-042f-4e13-925a-c1e40b77e010",
		Name:         "Canker Abomination",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.S82OldB522ChooseAnOpponentEntersWithMinusCounters("Canker Abomination"),
	})
}
