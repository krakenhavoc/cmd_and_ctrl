package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Starlit Soothsayer — Creature — Bat Cleric {2}{B}, 2/2 (Bloomburrow,
// #4467):
//
//	"Flying
//	 At the beginning of your end step, if you gained or lost life this
//	 turn, surveil 1. (Look at the top card of your library. You may put
//	 it into your graveyard.)"
//
// "If you gained or lost life this turn" is an intervening if (CR 603.4),
// read off the per-turn tallies Lunar Convocation reads as your end step
// begins. A loss counts whether it was damage, paid life or a "loses
// life" effect. The rule asks again as the trigger resolves, but those
// tallies only climb within a turn, so a condition true at the trigger
// is still true then, and the effect is the plain footprinted surveil
// (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0c0940c6-8126-4d58-809e-871cf8588ef9",
		Name:            "Starlit Soothsayer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, yourEndStepAndYouGainedOrLostLifeThisTurn,
				"Starlit Soothsayer — surveil 1 (you gained or lost life this turn)",
				Do(Surveil{N: 1})),
		},
	})
}

// yourEndStepAndYouGainedOrLostLifeThisTurn is "at the beginning of your
// end step, if you gained or lost life this turn", the condition's first
// check (CR 603.4).
func yourEndStepAndYouGainedOrLostLifeThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventBeginEndStep && ev.Actor == source.Controller &&
		youGainedOrLostLifeThisTurn(g, source.Controller)
}

// youGainedOrLostLifeThisTurn reads the player's life gained and life
// lost this turn.
func youGainedOrLostLifeThisTurn(g *game.Game, player uuid.UUID) bool {
	return b15LifeGainedThisTurn(g, player) > 0 || b18LifeLostThisTurn(g, player) > 0
}
