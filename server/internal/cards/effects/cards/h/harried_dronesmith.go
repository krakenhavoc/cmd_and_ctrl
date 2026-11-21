package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harried Dronesmith — Creature — Human Artificer {3}{R}, 2/3 (Murders at
// Karlov Manor, #4469):
//
//	"At the beginning of combat on your turn, create a 1/1 colorless
//	 Thopter artifact creature token with flying. It gains haste until
//	 end of turn. Sacrifice it at the beginning of your next end step."
//
// Lagomos, Hand of Hatred's beginning-of-combat token with Mardu Charm's
// keyed creation: the rest of the ability is the creation's continuation,
// told the tokens that were really made (a token doubler makes two, and
// both are "it"). Each gains haste until end of turn, and a delayed
// trigger (CR 603.7) sacrifices exactly those tokens at the beginning of
// their creator's next end step (ControllerTurnOnly). A token someone
// else has taken by then is not yours to sacrifice and stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6e2d8bd-8fe7-4de0-af25-7ad3e9dcb81b",
		Name:         "Harried Dronesmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Harried Dronesmith — create a 1/1 Thopter with haste; sacrifice it at your next end step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("1/1 colorless Thopter artifact with flying")), Count: 1}},
					}, ctx.ItemTail(harriedDronesmithHasteThenSacrifice))
				}),
		},
	})
}

// harriedDronesmithHasteThenSacrifice is "It gains haste until end of
// turn. Sacrifice it at the beginning of your next end step", over the
// tokens the creation made. An on-disk key.
var harriedDronesmithHasteThenSacrifice = effects.CardTail("harried-dronesmith/haste-then-sacrifice", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Harried Dronesmith — haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ScheduleDelayedTrigger{
		Label:              "Harried Dronesmith — sacrifice the Thopter",
		ControllerTurnOnly: true,
		Cards:              o.IDs,
		Body:               effects.SacrificeListedCardsBody,
	}.Apply(ctx)
})
