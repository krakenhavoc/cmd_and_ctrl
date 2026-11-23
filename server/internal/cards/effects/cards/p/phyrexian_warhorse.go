package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Warhorse — Creature — Phyrexian Horse {3}{B}, 3/3 (Dominaria
// United, #5487):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 When this creature enters, if it was kicked, create a 1/1 white
//	 Soldier creature token.
//	 {1}, Sacrifice another creature: This creature gets +2/+1 until end
//	 of turn."
//
// The pump is Dai Li Censor's sacrifice outlet: the creature is sacrificed
// as a cost, and a Warhorse that has left the battlefield (or become a new
// object) by resolution gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "05075332-6856-4147-a22d-3a5bd1e96b7b",
		Name:          "Phyrexian Warhorse",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 1},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Phyrexian Warhorse — kicked: create a 1/1 white Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature gets +2/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, ctx.Source()) {
					return nil
				}
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 1, Label: "Phyrexian Warhorse — +2/+1"}.Apply(ctx)
			},
		}},
	})
}
