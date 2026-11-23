package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Candlekeep Inspiration — Sorcery {4}{U} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Until end of turn, creatures you control have base power and toughness X/X,
//	 where X is the number of cards you own in exile and in your graveyard that
//	 are instant cards, are sorcery cards, and/or have an Adventure."
//
// Biomass Mutation's base-P/T set (layer 7b, the creatures fixed as it resolves,
// CR 611.2c) with X counted once then over the cards the caster owns in exile
// (face-down ones are not known to be anything) and in their graveyard, each
// card counted once however many of the three it is.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32d666ff-4d6b-4c60-8d3d-5bdae2d475fa",
		Name:         "Candlekeep Inspiration",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := candlekeepCount(ctx.Game, item.Controller)
			return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()),
				"Candlekeep Inspiration — base power and toughness X/X",
				game.SetBasePowerMod(x), game.SetBaseToughnessMod(x))
		},
	})
}

// candlekeepCount is the cards `player` owns in exile and in their graveyard
// that are instants, sorceries and/or have an Adventure.
func candlekeepCount(g *game.Game, player uuid.UUID) int {
	n := 0
	count := func(c game.Card) {
		if c.Owner == player && !c.FaceDown && (c.IsInstant() || c.IsSorcery() || effects.CardHasAdventure(c)) {
			n++
		}
	}
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			count(c)
		}
	}
	if p := g.PlayerByIDForEffect(player); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			count(c)
		}
	}
	return n
}
