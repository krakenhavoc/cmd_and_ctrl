package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Variable Chaser // Arc of Fortune — Creature — Human Wizard {2}{U},
// 2/3 // Sorcery {2}{U} (preparation card, CR 722):
//
//	"Flying, prowess
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Arc of Fortune — "Each player may discard their hand and draw seven
//	 cards."
//
// Every player chooses in APNAP order, then the players who said yes
// discard their hands, then each of them draws seven (CR 101.4,
// each_player_may_wheel.go).
//
// No simplification.
func init() {
	const id = "67c603cc-ad66-4a1c-8386-5901c9c01bfb"
	Register(Spec{
		OracleID:        id,
		Name:            "Variable Chaser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Arc of Fortune",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return arcOfFortuneMayWheel.Start(ctx)
		},
	})
}

var arcOfFortuneMayWheel = newMayWheel("confirm/arc-of-fortune-may-wheel",
	"Arc of Fortune — discard your hand and draw seven cards?")

// arcOfFortuneEmbark and arcOfFortuneAnswer are the spell's old shape, every
// player's question open at once and each yes wheeling at once. Kept, with
// their keys, for a restore point taken while one of those was open.
//
// arcOfFortuneEmbark is one player's discard and draw: the draw waits for a
// prompt the discards ask (a madness or a replacement offer; CR 608.2c, #4155).
// The player is the seat the answer was given for.
var arcOfFortuneEmbark = RegisterThenSteps("then-steps/arc-of-fortune-embark",
	Instruction(func(ctx *Context) error {
		_, err := discardWholeHand(ctx.Game, seatAtOccurrence(ctx))
		return err
	}),
	Instruction(func(ctx *Context) error {
		return ctx.Game.DrawNForEffect(seatAtOccurrence(ctx), 7)
	}),
)

// arcOfFortuneAnswer is a player's "yes": they discard their hand and draw
// seven. Keyed (#4291).
var arcOfFortuneAnswer = MayYes("confirm/arc-of-fortune-embark", func(ctx *Context, r game.ConfirmAnswer) error {
	return arcOfFortuneEmbark.RunAt(ctx, seatIndexOf(ctx.Game, r.Chooser))
})
