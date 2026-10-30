package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Everett K. Ross, Hapless Attaché — Legendary Creature — Human Advisor
// {2}{W}, 1/3 (Marvel Super Heroes Commander, #4455):
//
//	"Commander creatures you control get +1/+1 and have lifelink.
//	 Whenever an opponent attacks you with two or more creatures, draw a
//	 card."
//
// The static pair reads the "commander creature" flag on creatures you
// control (a commander that is not a creature right now gets nothing),
// +1/+1 in layer 7c and lifelink in layer 6. The trigger is Trouble in
// Pairs' attack shape: one trigger per declaration, however many
// creatures attack, and only when an opponent attacks YOU, not another
// player, with at least two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e017e323-e297-40de-bc77-dd49b610cf5a",
		Name:         "Everett K. Ross, Hapless Attaché",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Commander creatures you control get +1/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: msc7CommanderCreatureYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			{
				Label:     "Commander creatures you control have lifelink.",
				Layer:     game.Layer6Ability,
				AppliesTo: msc7CommanderCreatureYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = append(c.Abilities, "lifelink")
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentAttackedYouWithAtLeast(ev, source, g, 2, everettAttackLabel)
			}, everettAttackLabel, Do(DrawCards{N: 1}))),
		},
	})
}

const everettAttackLabel = "Everett K. Ross, Hapless Attaché — an opponent attacked you with two or more creatures: draw a card"

// msc7CommanderCreatureYouControl is the target test of "commander
// creatures you control".
func msc7CommanderCreatureYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.IsCommander && target.Controller == source.Controller
}
