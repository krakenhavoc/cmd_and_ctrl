package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyber Conversion — Instant for {U}{U}:
//
//	"Turn target creature face down. It's a 2/2 Cyberman artifact
//	 creature."
//
// The removal spell shape of #1209's primitive, and the only card in
// the ranked 6000-card roadmap that turns a permanent face down
// (batch 47, #454 — its skip line read "#656 face-down objects:
// Cyber Conversion"). No morph anywhere on it, so CR 708.7 answers
// plainly: nothing gave this permanent a way back up and nothing ever
// will. It stays a Cyberman until it leaves the battlefield — unless
// the card underneath has morph or disguise of its own (CR 702.37e).
//
// It targets ANY creature, tokens included. A token turned face down
// is not a rules problem — CR 111 has nothing to say about it, the
// token stays on the battlefield as a nameless 2/2, and CR 704.5d
// still removes it the moment it goes anywhere else.
//
// The second sentence is CR 708.2's LISTED characteristics, and since
// #1270 the engine has them: the body is CybermanBody(), which
// REPLACES CR 708.2a's default nameless 2/2 rather than decorating it.
// So the victim is an artifact (Shatter can hit it), a Cyberman (a
// Cyberman lord pumps it) and nothing else, and a Clone copying it
// copies the Cyberman (CR 708.2's second sentence). It shipped with
// the types declared as a caveat in #1209; the caveat is gone.
func init() {
	Register(Spec{
		OracleID:     "33761c1a-9848-45bf-b934-123cebab566b",
		Name:         "Cyber Conversion",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return turnSingleTargetFaceDownAs(CybermanBody)(ctx.Game, item)
		},
	})
}
