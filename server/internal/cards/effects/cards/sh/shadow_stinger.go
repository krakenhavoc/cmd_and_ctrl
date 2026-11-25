package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Stinger — Creature — Vampire Rogue {2}{B}, 1/4 (Zendikar Rising,
// S82 OLD #5508):
//
//	"Tap another untapped Rogue you control: This creature gains deathtouch
//	 until end of turn.
//	 Whenever this creature deals combat damage to a player, that player
//	 mills three cards."
//
// The tap is a cost paid with another Rogue (CR 602.2b); the Stinger never
// taps itself for it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "214562b1-fdf0-4d5d-93f0-594b31b80bb3",
		Name:         "Shadow Stinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Shadow Stinger — that player mills three cards",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					if ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
						return nil
					}
					return effects.MillCards{Player: ev.Target, N: 3}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap another untapped Rogue you control: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapAnotherUntapped("another untapped Rogue you control", effects.HasSubtype("Rogue")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"deathtouch"}, Label: "Shadow Stinger — deathtouch"}.Apply(ctx)
			},
		}},
	})
}
