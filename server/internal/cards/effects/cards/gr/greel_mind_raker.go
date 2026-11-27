package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greel, Mind Raker — Legendary Creature — Horror Spellshaper {3}{B}{B}, 3/3
// (Prophecy):
//
//	"{X}{B}, {T}, Discard two cards: Target player discards X cards at
//	 random."
//
// Mind Shatter's random discard (CR 701.9b) at the announced X, behind a cost
// of {X}{B}, {T} and two cards discarded by the activator (Zombie Infestation's
// discard). X = 0 discards nothing; a hand smaller than X is emptied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b84a10e-4c5c-407f-8b5b-a23d402f0529",
		Name:         "Greel, Mind Raker",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{B}, {T}, Discard two cards: Target player discards X cards at random.",
			Cost:    effects.Plus(effects.ManaCost("{X}{B}"), effects.TapCost(), effects.DiscardN(2, "two cards")),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if p := effects.TargetedPlayer(ctx); p != uuid.Nil {
					return effects.DiscardCards{Player: p, N: ctx.X()}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
