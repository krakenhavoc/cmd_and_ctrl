package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bruvac the Grandiloquent — Legendary Creature — Human Advisor
// {2}{U}, 1/4:
//
//	"If an opponent would mill one or more cards, they mill twice that
//	 many cards instead."
//
// The card the mill-amount seam was waiting on, and it has been named
// on the roadmap and in the sprint doc since S22 (#74, then #569). The
// engine could redirect a milled CARD long before it could double a
// mill: every milled card has gone through the shared exit primitive
// since #529, so Leyline of the Void and CR 903.9 both see one. What
// there was no event for is the NUMBER — "if an opponent would mill" is
// not a per-card zone move at all — and #569 built it
// (game.RepEventMill, one window per instruction, opened before
// anything leaves the library).
//
// So the whole card is one line. Bruvac's clause is symmetrical in the
// way that matters for CR 616.1: the replacement is HIS controller's
// and the affected player is the OPPONENT, so with a second
// mill-amount replacement on the same battlefield the ordering prompt
// goes to the player being milled, not to Bruvac's controller (#982).
//
// One event per instruction is also what makes "mill twice that many"
// mean what it says rather than "mill each card twice": an opponent
// told to mill three mills six, in one batch, with one set of mill
// payoffs firing over six cards.
//
// The count doubled is the one the instruction NAMED, not what the
// library holds. CR 701.13b makes a player mill as many as possible,
// so Bruvac doubling a twenty-card mill against a twelve-card library
// mills twelve — and doubles twenty, which is observable the moment a
// "plus N" shares the window.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "274b999f-f193-48fd-9a4a-0fdaf535e6c3",
		Name:         "Bruvac the Grandiloquent",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			OpponentsMillTwice("Bruvac the Grandiloquent — mill twice that many"),
		},
	})
}
