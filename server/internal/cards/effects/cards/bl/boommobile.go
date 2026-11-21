package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
// asks the colour first (`ChooseColorTail`) and then adds four mana of
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
	effects.Register(effects.Spec{
		OracleID:     "cb3edb86-ddd5-461c-98b5-ffb4df437a39",
		Name:         "Boommobile",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Boommobile — add four mana of any one color, spend only to activate abilities",
				boommobileMana),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Exhaust — {X}{2}{R}: This Vehicle deals X damage to any target. Put a +1/+1 counter on this Vehicle.",
				Exhaust: true,
				Cost:    effects.ManaCost("{X}{2}{R}"),
				Targets: effects.TargetAny(),
				// The counter waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						legal := ctx.LegalTargets()
						if len(legal) == 0 {
							return nil
						}
						return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: legal[0].ID, Amount: ctx.X()}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						if len(ctx.LegalTargets()) == 0 {
							return nil
						}
						return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}),
				),
			},
			{
				Label:   "Crew 2",
				Cost:    effects.CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Boommobile"),
			},
		},
	})
}

// boommobileMana is the ETB body: choose a colour, then add four mana of
// it that can only be spent to activate abilities.
func boommobileMana(g *game.Game, item *game.StackItem) error {
	effects.ChooseColorTail(game.ColorForMana, g, item.Controller, item.SourceCardID,
		"Boommobile — choose a color of mana to add",
		effects.NewContext(g, item).Tail(boommobileColorChosen))
	return nil
}

// boommobileColorChosen adds the four mana once the colour (the tail's Word)
// is known. An on-disk key.
var boommobileColorChosen = game.RegisterTail("then/boommobile-color-chosen", func(g *game.Game, o game.TailOutcome) error {
	return g.AddManaWithOptionsForEffect(o.Tail.Player, o.Tail.Source,
		"{"+o.Tail.Word+"4}",
		game.AddManaOptions{Restrictions: []string{effects.ManaRestrictActivate}})
})
