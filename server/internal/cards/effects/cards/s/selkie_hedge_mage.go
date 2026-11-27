package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selkie Hedge-Mage — Creature — Merfolk Wizard {2}{G/U} 2/2:
//
//	"When this creature enters, if you control two or more Forests, you may
//	 gain 3 life.
//	 When this creature enters, if you control two or more Islands, you may
//	 return target tapped creature to its owner's hand."
//
// Two triggers with an intervening "if" each (CR 603.4): the land count is
// checked as the creature enters and again as the trigger resolves. A land
// with both types counts toward both. The bounce's target (any tapped
// creature, yours included) is chosen as the trigger goes on the stack; the
// "you may" is asked then too (CR 603.5).
func init() {
	forests := func(g *game.Game, controller uuid.UUID) bool {
		return effects.CountControlled(g, controller, effects.IsLandWithSubtype("Forest")) >= 2
	}
	islands := func(g *game.Game, controller uuid.UUID) bool {
		return effects.CountControlled(g, controller, effects.IsLandWithSubtype("Island")) >= 2
	}
	enters := func(count func(*game.Game, uuid.UUID) bool) effects.When {
		return effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return count(g, source.Controller)
		})
	}
	effects.Register(effects.Spec{
		OracleID:     "1f1c369a-3209-405b-9880-0e10b86e5698",
		Name:         "Selkie Hedge-Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InterveningIf(effects.CondBoard{
				Name: "you control two or more Forests",
				Test: func(ctx *effects.Context) bool { return forests(ctx.Game, ctx.Controller()) },
			}, effects.On(game.EventETB, enters(forests), "Selkie Hedge-Mage — you may gain 3 life", effects.Do(effects.GainLife{Amount: 3}))),
				"Selkie Hedge-Mage — gain 3 life?"),
			effects.Optional(effects.Targeting(effects.InterveningIf(effects.CondBoard{
				Name: "you control two or more Islands",
				Test: func(ctx *effects.Context) bool { return islands(ctx.Game, ctx.Controller()) },
			}, effects.On(game.EventETB, enters(islands), "Selkie Hedge-Mage — you may return target tapped creature to its owner's hand", effects.S82Bro19BounceTarget)),
				effects.TargetCreature("target tapped creature", effects.TappedPermanent())),
				"Selkie Hedge-Mage — return target tapped creature to its owner's hand?"),
		},
	})
}
