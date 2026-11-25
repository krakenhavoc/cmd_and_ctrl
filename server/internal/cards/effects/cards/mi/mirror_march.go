package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror March — Enchantment {5}{R} (Ravnica Allegiance):
//
//	"Whenever a nontoken creature you control enters, flip a coin until you lose
//	 a flip. For each flip you won, create a token that's a copy of that
//	 creature. Those tokens gain haste. Exile them at the beginning of the next
//	 end step."
//
// Okaun's flip-until-you-lose chain, carrying the entering creature and the
// number of flips won. The first lost flip ends it and makes that many hasty
// copies in one creation, which the keyed tail exiles at the next end step.
// The copy reads the creature's last-known values if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "945eaa8b-0ba5-45a6-a02a-763e5bfe3af8",
		Name:         "Mirror March",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.NontokenCreatureEnteredUnderYourControl,
				"Mirror March — flip a coin until you lose a flip; for each flip you won, create a hasty token copy of that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if _, ok := ctx.TriggeringPermanent(); !ok {
						return nil
					}
					mirrorMarchFlip(g, item.Controller, item.SourceCardID, ctx.Trigger().Object.Ref().ID, 0, game.NewAskedItem(item))
					return nil
				}),
		},
	})
}

// mirrorMarchFlip queues one link of the chain.
func mirrorMarchFlip(g *game.Game, flipper, source, creature uuid.UUID, wins int, asked *game.AskedItem) {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  flipper,
		Source:   source,
		Question: "Mirror March — call the coin flip (flip until you lose)",
		Wins:     wins,
		Carry:    []uuid.UUID{creature},
		ThenKey:  mirrorMarchFlipped,
		Asked:    asked,
	})
}

// mirrorMarchFlipped is the settled call: a won flip asks the next, the first
// lost one makes the copies. An on-disk key, registered in init because its
// body queues itself.
var mirrorMarchFlipped game.CoinFlipThen

func init() {
	mirrorMarchFlipped = effects.CoinFlipThen("coin-flip/mirror-march-flip-until-you-lose", func(ctx *effects.Context, r game.CoinFlipped) error {
		creature := effects.CarriedID(r.Carry, 0)
		if effects.WonFirstFlip(r) {
			mirrorMarchFlip(ctx.Game, r.Flipper, r.Source, creature, r.Wins+1, game.NewAskedItem(r.Asked))
			return nil
		}
		if r.Wins == 0 {
			return nil
		}
		tmpl, ok := effects.TokenCopyTemplate(ctx.Game, creature)
		if !ok {
			return nil
		}
		effects.TokenCopyGainsHaste(&tmpl)
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(),
			Source:     ctx.Source(),
			Groups:     []game.TokenGroup{{Template: effects.CopiedToken(tmpl), Count: r.Wins}},
		}, ctx.ItemTail(mirrorMarchExile))
	})
}

var mirrorMarchExile = effects.SocB8ExileTokensAtNextEndStep("mirror-march/exile-the-copies",
	"Mirror March — exile the tokens")
