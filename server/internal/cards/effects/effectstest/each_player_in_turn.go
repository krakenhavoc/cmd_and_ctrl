package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// testMillEachHeard is what each run of testMillEachDone was told.
var TestMillEachHeard [][]uuid.UUID

// testMillEachDone records the cards a millEachPlayerInTurn chain milled.
var TestMillEachDone = game.RegisterTail("test/4253-mill-each-done", func(_ *game.Game, o game.TailOutcome) error {
	TestMillEachHeard = append(TestMillEachHeard, append([]uuid.UUID(nil), o.IDs...))
	return nil
})

// testExileTopEachHeard is what each run of testExileTopEachDone was told.
var TestExileTopEachHeard [][]uuid.UUID

// testExileTopEachDone records the cards an exileTopEachPlayerInTurn chain
// exiled.
var TestExileTopEachDone = game.RegisterTail("test/4253-exile-top-each-done", func(_ *game.Game, o game.TailOutcome) error {
	TestExileTopEachHeard = append(TestExileTopEachHeard, append([]uuid.UUID(nil), o.IDs...))
	return nil
})
