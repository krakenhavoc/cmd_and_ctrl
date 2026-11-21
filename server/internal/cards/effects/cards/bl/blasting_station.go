package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blasting Station — Artifact {3} (EDHREC rank 3477):
//
//	"{T}, Sacrifice a creature: This artifact deals 1 damage to any
//	 target.
//	 Whenever a creature enters, you may untap this artifact."
//
// The combo deck's sacrifice outlet. The activated ability is Goblin
// Bombardment's cost with a tap on top — Plus(TapCost(),
// SacrificeACreature()) — paid at announce, so the dies-triggers of
// the sacrificed creature land above the ability and resolve first;
// the damage is from the Station to whatever was announced, if it is
// still legal. The untap trigger fires for ANY creature entering —
// any player's, token or card — with a "you may" prompt, and it is
// what makes a persist creature a loop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a38d2d1-c4ff-4088-b1df-5feb9602ee2e",
		Name:         "Blasting Station",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: Blasting Station deals 1 damage to any target",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.B33DamageChosenTargetFromSource(1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33AnyCreatureEntered(ev, g)
			}, "Blasting Station — untap", effects.B33UntapSelf), "Blasting Station — untap it?"),
		},
	})
}
