package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haliya, Guided by Light — Legendary Creature — Human Soldier {2}{W},
// 3/3 (Edge of Eternities, #4461):
//
//	"Whenever Haliya or another creature or artifact you control enters,
//	 you gain 1 life.
//	 At the beginning of your end step, draw a card if you've gained 3 or
//	 more life this turn.
//	 Warp {W}"
//
// The first trigger is Weftstalker Ardent's creature-or-artifact entry
// watch with Haliya herself included. The end-step draw is not an
// intervening "if": the ability triggers every end step and checks your
// life gained this turn as it resolves (High Score's If, here over
// CondLifeGained), so life gained in response still counts. Warp is the
// shared alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9b0af3ad-1b4d-4480-af91-fe9baee5d1c5",
		Name:             "Haliya, Guided by Light",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && (c.IsCreature() || c.IsArtifact())
			}, "Haliya, Guided by Light — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
			effects.AtYourEndStep("Haliya, Guided by Light — draw a card if you've gained 3 or more life this turn",
				effects.Do(effects.If{Cond: effects.CondLifeGained{AtLeast: 3}, Then: []effects.Applier{effects.DrawCards{N: 1}}})),
		},
	})
}
