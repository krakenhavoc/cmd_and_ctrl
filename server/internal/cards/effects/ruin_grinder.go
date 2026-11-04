package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruin Grinder — Artifact Creature — Construct {5}{R}, 7/4:
//
//	"Menace
//	 When this creature dies, each player may discard their hand and
//	 draw seven cards.
//	 Mountaincycling {2}"
//
// The dies trigger is the shared "each player may" wheel: every player
// chooses in APNAP order, then the players who said yes discard their
// hands, then each of them draws seven (CR 101.4, each_player_may_wheel.go).
// Mountaincycling is the shared Typecycling constructor (any land with the
// Mountain type).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8d0445ef-370b-4f4f-8975-a40cb02c031e",
		Name:            "Ruin Grinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Ruin Grinder — each player may discard their hand and draw seven cards",
				func(g *game.Game, item *game.StackItem) error {
					return ruinGrinderMayWheel.Start(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}

var ruinGrinderMayWheel = newMayWheel("confirm/ruin-grinder-may-wheel",
	"Ruin Grinder — discard your hand and draw seven cards?")

// ruinGrinderWheel and its answer below are the trigger's old shape, every
// player's question open at once and each yes wheeling at once. Kept, with
// their keys, for a restore point taken while one of those was open.
//
// ruinGrinderWheel is one player's "discard your hand and draw seven cards".
// The steps find their player from the seat RunAt was given
// (ctx.Occurrence()). The draw waits for a prompt a discard asks (a
// replacement offer such as Library of Leng's; CR 608.2c, #4155).
var ruinGrinderWheel = RegisterThenSteps("then-steps/ruin-grinder-wheel",
	Instruction(func(ctx *Context) error {
		_, err := discardWholeHand(ctx.Game, seatAtOccurrence(ctx))
		return err
	}),
	Instruction(func(ctx *Context) error { return ctx.Game.DrawNForEffect(seatAtOccurrence(ctx), 7) }),
)

// The old answer is a player's "yes": they discard their hand and draw
// seven. Keyed (#4291).
var _ = MayYes("confirm/ruin-grinder-wheel", func(ctx *Context, r game.ConfirmAnswer) error {
	return ruinGrinderWheel.RunAt(ctx, seatIndexOf(ctx.Game, r.Chooser))
})
