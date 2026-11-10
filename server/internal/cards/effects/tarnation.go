package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tarnation — Plane — Thunder Junction (Unknown Event, #4870):
//
//	"Whenever a player commits a crime, they may draw a card. (Targeting
//	 opponents, anything they control, and/or cards in their graveyards is a
//	 crime.)
//	 Whenever chaos ensues, Tarnation deals 1 damage to any target."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability watches every crime announcement (CR 700.13), anyone's,
// and the player who committed it is the one asked and the one who draws.
// The chaos ability (CR 311.7) deals the damage with the plane as source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08081904-4115-4afd-ac82-f00421e54038",
		Name:         "Tarnation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, aPlayerCommittedACrime,
				"Tarnation — that player may draw a card", tarnationMayDraw),
			Targeting(WheneverChaosEnsues("Tarnation — 1 damage to any target", b13DamageToFirstTarget(1)),
				TargetAny()),
		},
	})
}

// aPlayerCommittedACrime is "whenever a player commits a crime" (CR 700.13):
// the once-per-announcement event, by anyone.
func aPlayerCommittedACrime(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBecomesTarget && ev.CrimeCommitted && ev.Actor != uuid.Nil
}

// tarnationMayDraw asks the player who committed the crime and draws for them.
func tarnationMayDraw(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	who := item.Trigger.Event.Actor
	ctx := NewContext(g, item)
	return MayChoice{
		Player:   who,
		Question: "Tarnation — draw a card?",
		YesLabel: "Draw a card",
		NoLabel:  "Don't draw",
		Draws:    1,
		Drawer:   who,
		Then:     tarnationDraw,
	}.Apply(ctx)
}

var tarnationDraw = MayYes("confirm/tarnation-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	return DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
})
