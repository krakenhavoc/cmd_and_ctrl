package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boommobile — Artifact — Vehicle {2}{R}{R}:
//
//	"When this Vehicle enters, add four mana of any one color. Spend
//	 this mana only to activate abilities.
//	 Exhaust — {X}{2}{R}: This Vehicle deals X damage to any target.
//	 Put a +1/+1 counter on this Vehicle. (Activate each exhaust
//	 ability only once.)
//	 Crew 2"
//
// The ETB is a trigger (it uses the stack, so it can be countered) that
// asks the colour first (`ChooseColorThen`) and then adds four mana of
// it carrying the activate-only spend restriction, so it cannot pay for
// a spell. The colour is asked separately because the any-colour picker
// (a pipe slot) cannot carry a spend restriction through its prompt. The mana can still pay the exhaust ability's
// {X}{2}{R}; Crew is paid with creatures, not mana. The exhaust ability is the engine's once-only
// exhaust marker (`Exhaust: true`), with a real announced X. The
// counter goes on after the damage, and only when the ability resolves:
// an illegal target fizzles the whole ability, as printed. Crew 2 is the
// ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb3edb86-ddd5-461c-98b5-ffb4df437a39",
		Name:         "Boommobile",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Boommobile — add four mana of any one color, spend only to activate abilities",
				boommobileMana),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Exhaust — {X}{2}{R}: This Vehicle deals X damage to any target. Put a +1/+1 counter on this Vehicle.",
				Exhaust: true,
				Cost:    ManaCost("{X}{2}{R}"),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					legal := ctx.LegalTargets()
					if len(legal) == 0 {
						return nil
					}
					if err := (DealDamage{Source: item.SourceCardID, Target: legal[0].ID, Amount: ctx.X()}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				},
			},
			{
				Label:  "Crew 2",
				Cost:   CrewCost(2),
				Effect: CrewEffect("Boommobile"),
			},
		},
	})
}

// boommobileMana is the ETB body: choose a colour, then add four mana of
// it that can only be spent to activate abilities.
func boommobileMana(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	ChooseColorThen(game.ColorForMana, g, controller, source,
		"Boommobile — choose a color of mana to add",
		func(g *game.Game, color string) error {
			return g.AddManaWithOptionsForEffect(controller, source,
				strings.Repeat("{"+color+"}", 4),
				game.AddManaOptions{Restrictions: []string{ManaRestrictActivate}})
		})
	return nil
}
