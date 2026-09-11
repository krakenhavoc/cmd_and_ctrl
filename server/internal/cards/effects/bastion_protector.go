package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bastion Protector — Creature — Human Soldier, {2}{W}, 3/3:
//
//	"Commander creatures you control get +2/+2 and have
//	 indestructible."
//
// A card that exists only in this format, and the cleanest test of
// whether the engine's commander model is real: the predicate asks
// `target.IsCommander`, a field that has been on `game.Card` since
// S02 and that until now almost nothing read.
//
// # Two statics, two layers
//
// The +2/+2 is layer 7c and the indestructible is layer 6. Same
// split as every "pump and grant" card in the catalog, and the same
// reason: a single `StaticAbility` carries one (Layer, SubLayer) and
// sorts into one bucket.
//
// Both share the predicate, which is factored out rather than
// written twice — if the two ever disagreed, the card would grant
// indestructible to a creature it had not pumped and nobody would
// notice for a sprint.
//
// # It can pump itself
//
// If Bastion Protector IS your commander, it is a commander creature
// you control and it gets its own +2/+2 and indestructible. That is
// the printed behaviour and it falls out of not excluding self. A
// partner pair gets both members.
//
// # Interaction with the 21-damage clock
//
// +2/+2 on a commander is worth more than +2/+2 on anything else,
// because CR 903.14a makes commander damage a second, much shorter
// life total. This card is why the S25 commander-damage rekey
// (per-commander rather than per-opponent keys) landed in the same
// sprint: with two partners both pumped, the two clocks have to be
// tracked separately or the maths on the board is wrong.
//
// No simplifications.
func init() {
	commanderCreatureYouControl := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() &&
			target.IsCommander &&
			target.Controller == source.Controller
	}
	Register(Spec{
		OracleID: "858f53ef-3fec-4aa0-867d-b7f040614b3c",
		Name:     "Bastion Protector",
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: commanderCreatureYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Layer:     game.Layer6Ability,
				AppliesTo: commanderCreatureYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = append(c.Abilities, "indestructible")
				},
			},
		},
	})
}
