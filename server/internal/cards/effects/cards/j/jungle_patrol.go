package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jungle Patrol — Creature — Human Soldier {3}{G}, 3/2 (Mirage, #5633):
//
//	"{1}{G}, {T}: Create a 0/1 green Wall creature token with defender named
//	 Wood.
//	 Sacrifice a token named Wood: Add {R}."
//
// The first ability makes the token (its name is part of the template). The
// second is a mana ability (CR 605.1a) whose cost sacrifices a token named
// Wood the activator controls (CR 605.1a lets a sacrifice be a mana
// ability's cost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bc8e918-7824-4ad7-a5c0-05537ce81641",
		Name:         "Jungle Patrol",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}: Create a 0/1 green Wall creature token with defender named Wood.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 green Wall with defender named Wood"), N: 1}),
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{SacrificeOther: effects.SacrificeN(1, "a token named Wood", effects.OldB813TokenNamedWood()).SacrificeOther},
			Produced: "{R}",
			Answers:  game.AnswerSacOutlet,
			Label:    "Sacrifice a token named Wood: Add {R}.",
		}},
	})
}
