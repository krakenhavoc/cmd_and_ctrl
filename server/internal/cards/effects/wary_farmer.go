package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wary Farmer — Creature — Kithkin Citizen {1}{G/W}{G/W}, 3/3 (Lorwyn
// Eclipsed, #4458):
//
//	"At the beginning of your end step, if another creature entered the
//	 battlefield under your control this turn, surveil 1."
//
// The condition is an intervening "if" (CR 603.4) on the turn's creature
// entry tally (Bristlebane Outrider's reading): checked as your end step
// begins, so a creature entering during the end step never makes it
// trigger (the card's ruling), and again as it resolves. "Another" needs
// a second entry when the Farmer itself entered this turn. The surveil is
// the footprinted Surveil primitive (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63eed1f5-442f-433d-989c-397b88543162",
		Name:         "Wary Farmer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "another creature entered the battlefield under your control this turn",
				Test: func(ctx *Context) bool {
					return waryFarmerAnotherCreatureEntered(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)
				},
			}, On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && waryFarmerAnotherCreatureEntered(g, source.Controller, source.InstanceID)
			}, "Wary Farmer — surveil 1", Do(Surveil{N: 1}))),
		},
	})
}

// waryFarmerAnotherCreatureEntered is "another
// creature entered the battlefield under your control this turn" for
// the permanent `source` controlled by `controller`: the turn's creature
// entry tally (Bristlebane Outrider's reading), needing one more entry
// when the source itself entered this turn. It takes ids rather than the
// card, so a resolving item whose source has left still answers.
//
// Caller must hold g.mu.
func waryFarmerAnotherCreatureEntered(g *game.Game, controller, source uuid.UUID) bool {
	need := 1
	if g.EnteredThisTurn(source) {
		need = 2
	}
	return g.EnteredWithCardTypeThisTurn(controller, "creature") >= need
}
