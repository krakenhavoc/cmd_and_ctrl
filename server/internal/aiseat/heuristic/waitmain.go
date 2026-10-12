package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"

// waitmain.go is the owner's gate on #2457, recorded in ADR 0126's
// amendment of 2026-10-12: the bot's own upkeep and draw step are not a
// spend window.
//
// ADR 0126 §5 lets a mana-and-taps move clear a lower bar in two windows
// because the mana it spends "costs nothing the bot would otherwise
// keep": mana empties between steps (CR 106.4), and a permanent tapped
// there untaps before the bot needs it again (CR 502.3). Before the
// bot's first main phase the premise is the other way round. A land
// tapped in its upkeep or draw step stays tapped through both of its
// main phases, so the mana is exactly the mana the turn's sorcery-speed
// plays would have used. Review game 2 (06e98afa, seq 248) shows the
// cost: Prismari Command cast in the draw step with all three lands left
// nothing for the main phase, where Mary Read and Anne Bonny was
// castable.
//
// The owner's rule: an instant cast before the bot's first main phase,
// with an empty stack, has to beat the best main-phase use of the same
// mana, or it waits for the main phase. The same cast is still on offer
// in the main phase (the card stays where it is, and nothing the bot can
// see changes between the draw step and the main phase), so the best
// main-phase use of that mana is never worth less than the cast is now.
// The cast waits. In the main phase it competes with the land drop, the
// turn plan (ADR 0136) and every sorcery-speed cast for the same mana.
//
// A move with something on the stack to answer is not held: the window
// is the empty stack only. Nor is an activated ability: the gate is the
// owner's, for casts.

// waitsForMain reports whether m is a cast held for the bot's main phase
// (Config.WaitForMain).
func (p *Policy) waitsForMain(st *state, m legal.Move) bool {
	return p.cfg.WaitForMain && st.beforeMyMain && m.Kind == legal.KindCast && p.instantSpeed(st, m)
}
