package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blot Out — Instant {2}{B} (March of the Machine: The Aftermath):
//
//	"Target opponent exiles a creature or planeswalker they control with
//	 the greatest mana value among creatures and planeswalkers they
//	 control."
//
// End of the Hunt at instant speed, so it shares End of the Hunt's
// resolution (s82MatB01ExileGreatestEdict): the targeted opponent is the
// one asked and chooses among the tied greatest; the permanent is not
// targeted, so hexproof and protection are irrelevant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9609cd5c-b537-4cb2-9dde-4b12795531f7",
		Name:         "Blot Out",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.S82MatB01ExileGreatestEdict(
			"Blot Out — exile a creature or planeswalker with the greatest mana value"),
	})
}
