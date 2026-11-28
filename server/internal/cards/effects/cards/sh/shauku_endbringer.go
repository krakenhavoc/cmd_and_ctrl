package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shauku, Endbringer — Legendary Creature — Vampire {5}{B}{B}, 5/5 (Mirage):
//
//	"Flying
//	 Shauku can't attack if there's another creature on the battlefield.
//	 At the beginning of your upkeep, you lose 3 life.
//	 {T}: Exile target creature and put a +1/+1 counter on Shauku."
//
// The attack restriction is read live at declaration over every other creature,
// whoever controls it (CR 508.1c). The upkeep life loss is unconditional. The
// tap ability exiles the target as it resolves if it is still legal (CR 608.2b)
// and then puts the counter on Shauku, only while it is still the permanent the
// ability came from (CR 400.7); the target may be Shauku itself, which then
// exiles itself and gets no counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "49478593-29ff-4e5a-b1e8-6462c2d6176d",
		Name:            "Shauku, Endbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.OldB804CantAttackIfAnotherCreature()},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Shauku, Endbringer — you lose 3 life", effects.Do(effects.YouLoseLife(3))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Exile target creature and put a +1/+1 counter on Shauku.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ExileFirstLegalCardTarget(g, item); err != nil {
					return err
				}
				return effects.PutCounterOnSelf(g, item)
			},
		}},
	})
}
