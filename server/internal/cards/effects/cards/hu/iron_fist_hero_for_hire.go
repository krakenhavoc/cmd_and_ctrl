package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Fist, Hero for Hire — Legendary Creature — Human Mercenary Hero {1}{R},
// 2/2 (Marvel Super Heroes Commander):
//
//	"Trample
//	 Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1
//	 until end of turn.)
//	 Power-up — {7}{R}: Iron Fist deals 5 damage divided as you choose among
//	 up to five targets. Put five +1/+1 counters on Iron Fist. (Activate each
//	 power-up ability only once. Reduce the cost by his mana cost if he entered
//	 this turn.)"
//
// Trample and prowess are printed keywords. The power-up is Bogardan Hellkite's
// divided damage on an activated ability: up to five targets, each assigned at
// least 1 of the 5 as the ability is put on the stack (CR 601.2d); a target
// that left takes nothing and its share is not redistributed (CR 608.2b). The
// five counters follow the damage. He deals the damage, so a counter-pumped
// Iron Fist does not change its size. The once-only rule and the discount for
// a power-up the turn he entered ({R} off the {7}{R}, CR 702.193b) are the
// engine's.
//
// No simplification.
func init() {
	powerUp := effects.PowerUp("{7}{R}: Iron Fist deals 5 damage divided as you choose among up to five targets. Put five +1/+1 counters on Iron Fist.",
		effects.ManaCost("{7}{R}"), func(g *game.Game, item *game.StackItem) error {
			if err := effects.DealDividedDamage(effects.NewContext(g, item)); err != nil {
				return err
			}
			return effects.PlusOneCountersOnThis(5)(g, item)
		})
	powerUp.Targets = effects.TargetAny().WithCount(0, 5).Dividing(effects.Divide(5))
	// The five counters pump him; the damage (announced with no target too,
	// "up to") removes: the union of what the row can do (ADR 0142).
	powerUp.Purpose.Answers = game.AnswerPump | game.AnswerRemove
	effects.Register(effects.Spec{
		OracleID:        "58c8bf6a-1b13-4c90-9e0b-24fb687f1e13",
		Name:            "Iron Fist, Hero for Hire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "prowess"},
		Activated:       []effects.ActivatedAbility{powerUp},
	})
}
