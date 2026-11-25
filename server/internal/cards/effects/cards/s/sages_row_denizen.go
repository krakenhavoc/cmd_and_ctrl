package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage's Row Denizen — Creature — Vedalken Wizard {2}{U}, 2/3 (Core Set 2020):
//
//	"Whenever another blue creature you control enters, target player mills two
//	 cards."
//
// Desperate Bloodseeker's mill with Anafenza's watch on another creature
// entering under your control, narrowed to a blue one: its colour is read as
// it is on the battlefield, through the layers. The player is the trigger's
// target, you included (CR 701.17a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc59aa9a-3abd-4479-a63e-ef0eeb35b5f2",
		Name:         "Sage's Row Denizen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.HasColor(source.TextColor("U"))
			}, "Sage's Row Denizen — target player mills two cards", effects.TargetPlayerMillsEffect(2)),
				effects.TargetPlayer("target player")),
		},
	})
}
