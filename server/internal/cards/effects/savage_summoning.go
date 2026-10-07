package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savage Summoning — Instant {G}:
//
//	"This spell can't be countered.
//	 The next creature spell you cast this turn can be cast as though it
//	 had flash. That spell can't be countered. That creature enters with
//	 an additional +1/+1 counter on it."
//
// A one-use promise about the next creature spell (#1852,
// game/next_spell_promise.go) carrying three riders: flash (read
// before the cast), can't be countered and the counter (marks on the
// spell, written as it becomes cast). The first creature spell cast
// spends the promise even if it was cast at sorcery speed or is then
// countered; a noncreature spell leaves it; it ends at cleanup.
//
// The spell's own "can't be countered" is the printed Spec flag.
func init() {
	Register(Spec{
		OracleID:        "30c09562-0bca-4382-ae84-aca00e0eb89b",
		Name:            "Savage Summoning",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantNextSpellPromise{From: "Savage Summoning", Promise: game.NextSpellPromise{
				Filter:          game.PermissionFilter{CreatureOnly: true},
				Flash:           true,
				CantBeCountered: true,
				CounterKind:     "+1/+1",
				Counters:        1,
				Text:            "The next creature spell you cast this turn can be cast as though it had flash, can't be countered, and enters with an additional +1/+1 counter.",
			}}.Apply(ctx)
		},
	})
}
