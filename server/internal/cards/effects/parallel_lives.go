package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Parallel Lives — Enchantment {3}{G} (EDHREC rank 519):
//
//	"If an effect would create one or more tokens under your control,
//	 it creates twice that many of those tokens instead."
//
// The green half of the token-doubling cycle, and the card the
// token-creation replacement event was built for (#762). One CR
// 701.7b event per creation INSTRUCTION, so "create three Saprolings"
// is one event that becomes six — not three events of two, which is
// the distinction that decides what a second doubler does.
//
// Two Parallel Lives are ×4 and ask nobody anything: they are two
// objects contributing one declared effect, which is the window
// #792's identical-modification skip exists for. A Parallel Lives
// beside an Anointed Procession is also ×4 and DOES prompt, because
// they are different declared effects and CR 616 gives the affected
// player the ordering even when the orderings agree.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "84dc94b2-95fb-4d53-aaa2-191cb645639f",
		Name:         "Parallel Lives",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			TokensDoubled("Parallel Lives: double tokens"),
		},
	})
}
