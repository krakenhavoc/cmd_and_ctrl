package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sure-Footed Infiltrator — Creature — Merfolk Rogue {3}{U}, 2/3 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Tap another untapped Rogue you control: This creature can't be blocked
//	 this turn.
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// The tap is a cost (CR 602.2b) on another untapped Rogue; the Infiltrator
// itself never pays it. The unblockability is a restriction until end of
// turn on the Infiltrator.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6108735-8370-4611-ad59-d7da7fd0323f",
		Name:         "Sure-Footed Infiltrator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Sure-Footed Infiltrator — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap another untapped Rogue you control: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapAnotherUntapped("another untapped Rogue you control", effects.HasSubtype("Rogue")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked, Label: "Sure-Footed Infiltrator — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}
