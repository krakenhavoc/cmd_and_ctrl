package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ninja Teen — Enchantment — Class {2}{B} (Teenage Mutant Ninja Turtles,
// #4457, #4680):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Whenever a creature you control leaves the battlefield, each
//	 opponent loses 1 life.
//	 {1}{B}: Level 2
//	 Creatures you control get +1/+0 and have menace.
//	 {B}: Level 3
//	 Creature cards in your graveyard have sneak {3}{B}.
//	 You may cast creature spells from your graveyard using their sneak
//	 abilities."
//
// Level 1 is The Ozolith's trigger condition (aCreatureYouControlLeft):
// any departure of a creature its controller controlled as it left, read
// from last-known information (CR 603.10a), tokens included, then
// Blood Artist's "each opponent loses 1 life". Level 2 is Jetmir's two
// statics, the layer 7c +1/+0 and the layer 6 menace, gated at Level(2)
// so they stay at level 3 (CR 716.2a).
//
// Level 3 is the granted sneak (#4680, game/sneak.go): one standing
// permission over the graveyard, gated at Level(3), that gives each
// creature card there sneak {3}{B} and lets a creature spell be cast from
// there only by a sneak cost, the granted one or one the card prints. The
// sneak window, the return of an unblocked attacker and the tapped and
// attacking entry are sneak's own (CR 702.190a-b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     NinjaTeenOracle,
		Name:         "Ninja Teen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{1}{B}")),
			effects.LevelUp(3, effects.ManaCost("{B}")),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.ACreatureYouControlLeft, "Ninja Teen — each opponent loses 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
		Static: []game.StaticAbility{
			{
				Label:      "Creatures you control get +1/+0.",
				Layer:      game.Layer7PT,
				SubLayer:   game.SubLayer7C_Modify,
				AppliesTo:  effects.CreaturesYouControl,
				ActiveWhen: effects.Level(2),
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
				},
			},
			{
				Label:      "Creatures you control have menace.",
				Layer:      game.Layer6Ability,
				AppliesTo:  effects.CreaturesYouControl,
				ActiveWhen: effects.Level(2),
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasAbility(c.Abilities, "menace") {
						c.Abilities = append(c.Abilities, "menace")
					}
				},
			},
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: effects.CreatureCardsInYourGraveyardHaveSneak("{3}{B}", "Ninja Teen"),
			ActiveWhen: effects.Level(3),
		}},
	})
}

const NinjaTeenOracle = "2c5044ff-a303-4276-8e23-7caac30a275b"
