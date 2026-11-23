package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mephidross Vampire — Creature — Vampire {4}{B}{B}, 3/4 (Special Guests):
//
//	"Flying
//	 Each creature you control is a Vampire in addition to its other
//	 creature types and has "Whenever this creature deals damage to a
//	 creature, put a +1/+1 counter on this creature.""
//
// Two statics on one source. The type is a layer 4 addition (CR 613.1d) to
// every creature its controller controls, itself included. The trigger is a
// layer 6 grant of a catalog bundle (ADR 0093), so "this creature" is each
// recipient: one trigger per damage event to a creature, combat or not, a
// counter on the recipient only, and a creature that stops being yours or
// loses the Vampire loses both.
//
// No simplification.
const mephidrossVampireGrant = "mephidross-vampire/counter-on-damage"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "a6763205-47c7-4d7f-be8c-2f69e8321048",
		Name:            "Mephidross Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key: mephidrossVampireGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventDealDamage, effects.ThisDealtDamageToACreature,
					"Mephidross Vampire — put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			},
			Text: "Whenever this creature deals damage to a creature, put a +1/+1 counter on this creature.",
		}},
		Static: []game.StaticAbility{
			{
				Label:     "Each creature you control is a Vampire in addition to its other creature types.",
				Layer:     game.Layer4Type,
				AppliesTo: mephidrossVampireYourCreatures,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Subtypes, "Vampire") {
						c.Subtypes = append(c.Subtypes, "Vampire")
					}
				},
			},
			effects.GrantAbilities(mephidrossVampireYourCreatures, mephidrossVampireGrant),
		},
	})
}

// mephidrossVampireYourCreatures is "each creature you control", read as
// the layers have it, the Vampire itself included.
func mephidrossVampireYourCreatures(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller
}
