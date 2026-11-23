package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// search_scry_tail_restore_test.go — #4253 slice 4: the "then" of a library
// search and of a scry or surveil is a registered tail (keyed_tails.go), so
// a table waiting on a fetched permanent's entry question, or on a scry, is
// a restore point. Each test takes a STRICT restore (the boot path) while
// the prompt is open, answers it on the restored table, and checks that the
// card still finishes: the tail ran off the disk, not off a closure.
//
// The fetched permanent's question is a Basic Land with a shockland's entry
// clause (basicShockland), a Clone's copy question (a creature search), or a
// Kismet and a Thalia's CR 616 ordering prompt (an artifact creature).

// basicShockland is a Basic Land card that is also Blood Crypt: a search
// for a basic land may take it, and it enters with the "pay 2 life?"
// question.
func BasicShockland() game.Card {
	return game.Card{Name: "Blood Crypt", TypeLine: "Basic Land — Swamp Mountain", OracleID: BloodCryptOracle}
}

// answerCloneCopy answers the open copy question by copying `bear`.
func AnswerCloneCopy(t *testing.T, g *game.Game, bear uuid.UUID) {
	t.Helper()
	c := CopyPrompt(g)
	if c == nil {
		t.Fatal("no copy question is open")
	}
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, bear); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
}
