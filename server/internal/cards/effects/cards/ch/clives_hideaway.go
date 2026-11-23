package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clive's Hideaway — Land — Town:
//
//	"Hideaway 4
//	 {T}: Add {C}.
//	 {2}, {T}: You may play the exiled card without paying its mana cost
//	 if you control four or more legendary creatures."
//
// The Lorwyn hideaway lands' shape (hideaway_lands.go) with two printed
// differences: it does not enter tapped, and the activation costs {2}.
// The condition is part of the effect, asked as the ability resolves, over
// the creatures you control that are legendary now.
func init() {
	const name = "Clive's Hideaway"
	effects.Register(effects.Spec{
		OracleID:     "283f743f-6e79-49de-b7ed-08e6ffb64cc6",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Hideaway(name, 4)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: You may play the exiled card without paying its mana cost if you control four or more legendary creatures.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !controlsAtLeastLegendaryCreatures(g, item.Controller, 4) {
					return nil
				}
				return effects.PlayHiddenCard{
					Source: effects.HiddenRefOfActivation(item),
					Label:  name + " — play the exiled card without paying its mana cost",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// controlsAtLeastLegendaryCreatures is "you control N or more legendary
// creatures", read through the layered view.
func controlsAtLeastLegendaryCreatures(g *game.Game, controller uuid.UUID, n int) bool {
	count := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.IsLegendary() {
			count++
		}
	}
	return count >= n
}
