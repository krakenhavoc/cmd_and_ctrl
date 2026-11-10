package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Graaz, Unstoppable Juggernaut — Legendary Artifact Creature —
// Juggernaut {8}, 7/5:
//
//	"Juggernauts you control attack each combat if able.
//	 Juggernauts you control can't be blocked by Walls.
//	 Other creatures you control have base power and toughness 5/3 and are
//	 Juggernauts in addition to their other creature types."
//
// Three statics on the board. The creature-type grant is layer 4 and the
// base P/T is layer 7b, so the new Juggernauts are read after the type
// change (CR 613.1d, CR 613.4b). The attack requirement is Goblin
// Rabblemaster's AttacksEachCombatWhere and the Wall bar is Juggernaut's
// pair rule, both scoped to Juggernauts its controller controls.
//
// No simplification.
func init() {
	isMyJuggernaut := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller &&
			target.HasSubtype(source.TextSubtype("Juggernaut"))
	}
	isOtherCreatureOfMine := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
	}
	atk := AttacksEachCombatWhere(isMyJuggernaut)
	atk.Label = "Juggernauts you control attack each combat if able."
	Register(Spec{
		OracleID:     "aebfdcc0-c5b9-4737-8498-764a5021b087",
		Name:         "Graaz, Unstoppable Juggernaut",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			atk,
			{
				Label:     "Other creatures you control are Juggernauts in addition to their other creature types.",
				Layer:     game.Layer4Type,
				AppliesTo: isOtherCreatureOfMine,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					j := source.TextSubtype("Juggernaut")
					if !slices.Contains(c.Subtypes, j) {
						c.Subtypes = append(c.Subtypes, j)
					}
				},
			},
			{
				Label:     "Other creatures you control have base power and toughness 5/3.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: isOtherCreatureOfMine,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 5, 3
				},
			},
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnMatching(And(Subtype("Juggernaut"), Creature(), YouControl())), Subtype("Wall"), "Walls"),
		},
	})
}
