package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kyoshi Warrior Exemplars — Creature — Human Warrior Ally {3}{G}, 4/3
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever this creature attacks, if you control eight or more lands,
//	 creatures you control get +2/+2 until end of turn."
//
// An intervening "if" (CR 603.4): the land count is checked as the
// creature is declared as an attacker and again as the ability resolves.
// The creatures that get the bonus are fixed as it resolves (CR 611.2c),
// so a creature that enters afterwards doesn't get it.
//
// No simplification.
func init() {
	const label = "Kyoshi Warrior Exemplars — creatures you control get +2/+2 until end of turn"
	eightLands := ControlsAtLeast(8, MatchLand)
	Register(Spec{
		OracleID:     "4bb0596f-df1d-4977-bb8b-4d2e6648dcb3",
		Name:         "Kyoshi Warrior Exemplars",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control eight or more lands",
				Test: func(ctx *Context) bool { return eightLands(ctx.Game, ctx.Controller(), ctx.Source()) },
			}, On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ThisAttacked(ev, source, lki, g) && eightLands(g, source.Controller, source.InstanceID)
			}, label, Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Toughness: 2, Label: label}))),
		},
	})
}
