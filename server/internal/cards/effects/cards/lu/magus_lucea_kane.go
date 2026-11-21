package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magus Lucea Kane — Legendary Creature — Human Tyranid Wizard {1}{G}{U}{R},
// 1/1 (Warhammer 40,000 Commander, S82 GAP batch 193, #5948):
//
//	"Spiritual Leader — At the beginning of combat on your turn, put a +1/+1
//	 counter on target creature.
//	 Psychic Stimulus — {T}: Add {C}{C}. When you next cast a spell with {X}
//	 in its mana cost or activate an ability with {X} in its activation cost
//	 this turn, copy that spell or ability. You may choose new targets for the
//	 copy. (A copy of a permanent spell becomes a token.)"
//
// The combat trigger's target is any creature. The mana ability's rider
// schedules a delayed trigger (Brass Infiniscope's shape) watching both a cast
// of a spell whose own mana cost has an {X} and the activation of a stack
// ability whose cost has an {X}; it copies whichever fired it, from last-known
// information, and offers new targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae9f5c80-bc96-4ab3-bb5b-e8bd470e9eab",
		Name:         "Magus Lucea Kane",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Magus Lucea Kane — a +1/+1 counter on target creature", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "{T}: Add {C}{C}. When you next cast a spell with {X} in its mana cost or activate an ability with {X} in its activation cost this turn, copy that spell or ability",
			Rider:    effects.GapB193ScheduleCopyNextXThing("Magus Lucea Kane"),
		}},
	})
}
