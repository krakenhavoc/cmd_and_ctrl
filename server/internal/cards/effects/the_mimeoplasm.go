package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Mimeoplasm — Legendary Creature — Ooze {2}{B}{G}{U}, 0/0 (Outlaws of
// Thunder Junction Commander, #5457):
//
//	"As The Mimeoplasm enters, you may exile two creature cards from
//	 graveyards. If you do, it enters as a copy of one of those cards with
//	 a number of additional +1/+1 counters on it equal to the power of the
//	 other card."
//
// Body Double's enters-as-a-copy of a graveyard card (CR 707.2, CR
// 614.1c) over two cards (game.CopySelector.ExileWithAnother). The
// copy_target prompt is asked twice while the entry is paused: first the
// card to copy, then the other card, among the creature cards in every
// graveyard. Both are exiled as The Mimeoplasm enters, it takes the first
// card's printed values, and the +1/+1 counters, equal to the second
// card's power as it sat in the graveyard (its own characteristic-
// defining ability included, CR 604.3), are counters it enters with, so
// an effect that changes those sees them. Nothing is targeted, so nothing
// that stops targeting narrows the choice (CR 115.10a).
//
// Declining either question, or fewer than two creature cards in all
// graveyards, exiles nothing: The Mimeoplasm enters as its printed 0/0
// and dies (CR 704.5f).
//
// No simplification.
func init() {
	r := EntersAsCopyOf("The Mimeoplasm", anyCreatureCardInAGraveyard, nil)
	r.RowLabel = "May exile two creature cards from graveyards and enter as a copy of one, with +1/+1 counters equal to the other's power."
	r.PromptQuestion = "The Mimeoplasm — exile two creature cards from graveyards?"
	r.CopySelector.ExileWithAnother = true
	Register(Spec{
		OracleID:     "55123455-d435-46d4-b0cd-0c1614343be2",
		Name:         "The Mimeoplasm",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{r},
	})
}
