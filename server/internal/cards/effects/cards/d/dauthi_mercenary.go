package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauthi Mercenary — Creature — Dauthi Knight Mercenary {2}{B}, 2/1
// (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 {1}{B}: This creature gets +1/+0 until end of turn."
//
// Shadow is a printed keyword. The pump is Shivan Dragon's at {1}{B},
// with the shared thisGetsUntilEndOfTurn body: no tap in the cost, so it
// can be activated any number of times, summoning sick or not, and the
// boost ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b57c480-e2c5-44cd-820f-e208041bfd18",
		Name:            "Dauthi Mercenary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Dauthi Mercenary — +1/+0"),
		}},
	})
}
