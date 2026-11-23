package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirrorshell Crab — Artifact Creature — Crab {5}{U}{U}, 5/7 (Neon Dynasty,
// #5493):
//
//	"Ward {3}
//	 Channel — {2}{U}, Discard this card: Counter target spell or ability
//	 unless its controller pays {3}."
//
// Channel is an activated ability that works from the hand (CR 207.2c); the
// discard is a cost, so a countered channel still costs the card. The
// counter is the shared "unless its controller pays" prompt, which asks the
// stack item's controller as the ability resolves and handles an ability on
// the stack as well as a spell (CR 701.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd727e02-5064-4b89-9398-e3cc4463ba56",
		Name:         "Mirrorshell Crab",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Mirrorshell Crab — ward {3}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {2}{U}, Discard this card: Counter target spell or ability unless its controller pays {3}.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetSpellOrAbility("target spell or ability"),
			Effect:  effects.CounterTheTargetSpellUnlessPaid("Mirrorshell Crab", "{3}"),
		}},
	})
}
