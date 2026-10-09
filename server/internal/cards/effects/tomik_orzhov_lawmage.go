package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tomik, Orzhov Lawmage — Legendary Creature — Human Advisor {1}{W}, 2/1
// (Reality Fracture, tracker #2795):
//
//	"Flying
//	 Planeswalkers you control have "No more than one creature can attack
//	 this planeswalker each combat."
//	 {T}: Target creature with a +1/+1 counter on it gains flying until end
//	 of turn."
//
// The granted line is an ADR 0093 bundle carrying The Eternal Wanderer's
// attack limit (#2821): each planeswalker you control gets its own limit
// of one, counted against itself, and loses it the moment Tomik leaves or
// it changes control.
func init() {
	Register(Spec{
		OracleID:        "7a501f7e-eec8-45f7-9ac3-483fd8f5ca5e",
		Name:            "Tomik, Orzhov Lawmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key:          tomikAttackLimitGrant,
			AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackThisEachCombat(1)},
			Text:         "No more than one creature can attack this planeswalker each combat.",
		}},
		Static: []game.StaticAbility{
			GrantAbilitiesToYourPlaneswalkers(tomikAttackLimitGrant),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature with a +1/+1 counter on it gains flying until end of turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with a +1/+1 counter on it", tomikHasPlusCounter),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"flying"},
						Label:    "Tomik, Orzhov Lawmage — gains flying until end of turn",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// tomikAttackLimitGrant is the bundle every planeswalker Tomik's
// controller controls carries.
const tomikAttackLimitGrant = "tomik-orzhov-lawmage/one-attacker"

// tomikHasPlusCounter is "with a +1/+1 counter on it".
func tomikHasPlusCounter(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.Counters[game.CounterPlusOne] > 0
}
