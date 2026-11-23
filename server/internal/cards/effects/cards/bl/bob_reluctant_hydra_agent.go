package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bob, Reluctant HYDRA Agent — Legendary Creature — Human Coward Villain
// {1}{B}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Bob attacks alone, return him to his owner's hand. If you do,
//	 each opponent loses 2 life and you gain 2 life."
//
// "Alone" is game.AttackedAlone read when the trigger is harvested, on Bob's
// own attack. The bounce is Mirkwood Nurturer's: BounceToHand with a
// registered "if you do" tail, which runs the drain only when Bob actually
// reached his owner's hand (a commander sent to the command zone instead gets
// no drain, CR 903.9; Bob removed in response does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9caa490e-dc28-4069-9ffb-4b1afd8fdd54",
		Name:         "Bob, Reluctant HYDRA Agent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
				return effects.ThisAttacked(ev, source, c, g) && game.AttackedAlone(g)
			}, "Bob — return him to his owner's hand; if you do, each opponent loses 2 life and you gain 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BounceToHand{Target: item.SourceCardID, ThenTail: ctx.ItemTail(bobDrainTail)}.Apply(ctx)
				}),
		},
	})
}

var bobDrainTail = effects.CardTail("bob-reluctant-hydra-agent/drain", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	if err := (effects.EachOpponentLosesLife{N: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Amount: 2}.Apply(ctx)
})
