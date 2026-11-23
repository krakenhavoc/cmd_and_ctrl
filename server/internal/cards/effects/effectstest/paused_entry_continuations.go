package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_entry_continuations_test.go — #478. A battlefield ENTRY that
// queues a prompt used to lose the card.
//
// The repro is the issue's: an opponent controls Kismet and Thalia,
// Heretic Cathar, two "enters tapped" replacements that both apply to
// your Guildgate. One of them is a silent modification; TWO of them is
// a CR 616 ordering prompt, and the search bailed on that prompt with
// the card still in the library. The player answered a question that
// did nothing, the permanent never arrived, and the library was never
// shuffled either.
//
// The fix gives the entry the same posture the exit has had since #529:
// nothing moves while the question is open, and the SAME resume
// finishes the move and runs what the effect still owed — the shuffle,
// EventSearchLibrary and the caller's Then for a search, the CR 400.7
// new object for an exile return.

// pe2EntersTapped seats both enters-tapped hatebears under `opp`, so
// any nonbasic land or creature entering under anybody else gets two
// applicable replacements and therefore a CR 616 ordering prompt.
func Pe2EntersTapped(g *game.Game, opp uuid.UUID) {
	B12Push(g, opp, "Kismet", "Enchantment", KismetOracle, 0, 0)
	B12Push(g, opp, "Thalia, Heretic Cathar", "Legendary Creature — Human Soldier",
		B13ThaliaHereticCatharOracle, 3, 2)
}

// peAnswerOrder answers an open CR 616 ordering prompt for `chooser` in
// the order it was offered.
func PeAnswerOrder(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := B06ReplacementOrderFor(g, chooser)
	if c == nil {
		t.Fatalf("no replacement-order prompt for %s", chooser)
	}
	if err := g.ResolveReplacementOrder(c.ID, chooser, c.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}
