package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciless Harlequin — Creature — Human Assassin {2}{B}, 2/1 (Assassin's
// Creed, #5948):
//
//	"Freerunning {1}{B} (You may cast this spell for its freerunning cost
//	 if you dealt combat damage to a player this turn with an Assassin or
//	 commander.)
//	 When this creature enters, you draw a card and you lose 1 life."
//
// Freerunning is the Brotherhood Ambushers alternative cost; the enters
// trigger is a draw then a life loss, with no choice in either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "10e0ccdf-9a31-402a-9fbf-f501ce63fa52",
		Name:             "Merciless Harlequin",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Freerunning("{1}{B}", "Assassin")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Merciless Harlequin — you draw a card and you lose 1 life",
				Do(DrawCards{N: 1}, youLoseLife(1))),
		},
	})
}
