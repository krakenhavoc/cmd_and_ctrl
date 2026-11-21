package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Blue Spirit — Legendary Creature — Human Rogue Ally {3}{U}, 2/4
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"You may cast the first creature spell you cast each turn as though
//	 it had flash.
//	 Whenever a nontoken creature you control enters during combat, draw
//	 a card."
//
// The flash half is Radagast of Rhosgobel's derived cast-timing rule
// (CastFirstCreatureSpellAsThoughFlash), gated on the per-turn cast
// tally, so the second creature spell of a turn finds the window shut,
// and it stops when The Blue Spirit leaves.
//
// "During combat" is the combat phase (CR 506.1: beginning of combat
// through end of combat), read as the creature enters. The Blue Spirit
// itself is a nontoken creature you control, so flashing it in during
// combat draws a card too. A token does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "497e9c7c-a16d-43e5-a7ec-04153aa2566a",
		Name:         "The Blue Spirit",
		Completeness: effects.CompletenessFull,
		CastTimings: []game.CastTimingRule{
			effects.CastFirstCreatureSpellAsThoughFlash("You may cast the first creature spell you cast each turn as though it had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, avB30dNontokenCreatureYouControlEnteredDuringCombat,
				"The Blue Spirit — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// avB30dNontokenCreatureYouControlEnteredDuringCombat is "whenever a
// nontoken creature you control enters during combat", the source
// itself included.
func avB30dNontokenCreatureYouControlEnteredDuringCombat(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && !effects.IsToken(c) && g.Turn.Phase == game.PhaseCombat
}
