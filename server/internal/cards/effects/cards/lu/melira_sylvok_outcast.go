package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Melira, Sylvok Outcast — Legendary Creature — Human Scout {1}{G}, 2/2:
//
//	"You can't get poison counters.
//	 Creatures you control can't have -1/-1 counters put on them.
//	 Creatures your opponents control lose infect."
//
// The first two lines cancel the placement (CR 101.2) before any other
// replacement can change it; the third strips infect through the layered
// keyword view, so an opponent's creature deals its damage as ordinary damage.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6f72f15-d103-49fb-a808-054b278d00a8",
		Name:         "Melira, Sylvok Outcast",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LoseKeywords(effects.B27CreaturesOpponentsControl, "infect"),
		},
		Replacements: []game.ReplacementEffect{
			effects.CantGetPoisonCountersReplacement("Melira, Sylvok Outcast: you can't get poison counters"),
			effects.CreaturesYouControlCantHaveMinusOneCounters("Melira, Sylvok Outcast: creatures you control can't have -1/-1 counters put on them"),
		},
	})
}
