package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "8d0445ef-370b-4f4f-8975-a40cb02c031e",
		Name:            "Ruin Grinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Ruin Grinder — each player may discard their hand and draw seven cards",
				func(g *game.Game, item *game.StackItem) error {
					return ruinGrinderMayWheel.Start(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}

var ruinGrinderMayWheel = effects.NewMayWheel("confirm/ruin-grinder-may-wheel",
	"Ruin Grinder — discard your hand and draw seven cards?")

// ruinGrinderWheel and its answer below are the trigger's old shape, every
// player's question open at once and each yes wheeling at once. Kept, with
// their keys, for a restore point taken while one of those was open.
//
// ruinGrinderWheel is one player's "discard your hand and draw seven cards".
// The steps find their player from the seat RunAt was given
// (ctx.Occurrence()). The draw waits for a prompt a discard asks (a
// replacement offer such as Library of Leng's; CR 608.2c, #4155).
var ruinGrinderWheel = effects.RegisterThenSteps("then-steps/ruin-grinder-wheel",
	effects.Instruction(func(ctx *effects.Context) error {
		_, err := effects.DiscardWholeHand(ctx.Game, effects.SeatAtOccurrence(ctx))
		return err
	}),
	effects.Instruction(func(ctx *effects.Context) error { return ctx.Game.DrawNForEffect(effects.SeatAtOccurrence(ctx), 7) }),
)

// The old answer is a player's "yes": they discard their hand and draw
// seven. Keyed (#4291).
var _ = effects.MayYes("confirm/ruin-grinder-wheel", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	return ruinGrinderWheel.RunAt(ctx, effects.SeatIndexOf(ctx.Game, r.Chooser))
})
