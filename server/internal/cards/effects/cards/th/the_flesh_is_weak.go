package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Flesh Is Weak — Enchantment {2}{W}{U}{B} (Warhammer 40,000
// Commander):
//
//	"When this enchantment enters, put a +1/+1 counter on each creature
//	 you control.
//	 Creatures you control with +1/+1 counters on them are artifacts in
//	 addition to their other types.
//	 Nonartifact creatures get -1/-1."
//
// The enters trigger counts the creatures as it resolves and puts the
// counters as one placement each (B11PutCounterOnEachCreatureYouControl).
// The type change is a layer 4 static (CR 613.1d) that reads the counters
// on each creature; the -1/-1 is a layer 7c static (CR 613.4c) on EVERY
// creature, yours and your opponents', that is not an artifact once layer
// 4 has run, so a creature carrying a +1/+1 counter becomes an artifact
// and is not shrunk by it (CR 613.1, 613.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dafbb927-3d9d-4be7-b427-f6564b295161",
		Name:         "The Flesh Is Weak",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("The Flesh Is Weak — put a +1/+1 counter on each creature you control",
				effects.Do(effects.Instruction(effects.PutAPlusOneCounterOnEachCreatureYouControlStep))),
		},
		Static: []game.StaticAbility{
			{
				Layer: game.Layer4Type,
				Label: "Creatures you control with +1/+1 counters on them are artifacts in addition to their other types.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.IsCreature() && target.Controller == source.Controller &&
						target.Counters[game.CounterPlusOne] > 0
				},
				Apply: effects.S82OneB18AddArtifactType,
			},
			{
				Label:    "Nonartifact creatures get -1/-1.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
					return target.IsCreature() && !target.IsArtifact()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power--
					c.Toughness--
				},
			},
		},
	})
}
