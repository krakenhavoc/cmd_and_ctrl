package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plaza of Harmony — Land (Ravnica Allegiance):
//
//	"When this land enters, if you control two or more Gates, you gain 3 life.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any type that a Gate you control could produce."
//
// The Plaza is not a Gate itself, so it never counts toward its own life gain
// or its own mana. The life gain is an intervening if (CR 603.4); the second
// mana ability is Reflecting Pool's derived mana (CR 106.7) over the Gates you
// control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ff1d6d8-8cea-4a25-90d9-b575f4c99bc8",
		Name:         "Plaza of Harmony",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: effects.ManaAbilityCost{Tap: true},
				DerivedMatch: func(c game.Card, controller uuid.UUID) bool {
					return c.Controller == controller && c.IsLand() && c.HasSubtype("Gate")
				},
				DerivedColorsOnly:       false,
				DerivesFromOtherSources: true,
				Label:                   "Add one mana of any type that a Gate you control could produce",
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control two or more Gates",
				Test: func(ctx *effects.Context) bool {
					return len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.HasSubtype("Gate"))) >= 2
				},
			}, effects.WhenThisEnters("Plaza of Harmony — you gain 3 life", effects.Do(effects.GainLife{Amount: 3}))),
		},
	})
}
