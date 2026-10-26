package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summoner's Sending — Enchantment {1}{W}:
//
//	"At the beginning of your end step, you may exile target creature
//	 card from a graveyard. If you do, create a 1/1 white Spirit
//	 creature token with flying. Put a +1/+1 counter on it if the
//	 exiled card's mana value is 4 or greater."
//
// The "you may" is the trigger's CR 603.5 prompt, asked before the
// target. "If you do" hangs off the exile's continuation: a card that
// left the graveyard in response, or that a replacement kept out of
// exile, makes no Spirit. The mana value is the exiled card's, read
// once it is in exile, and the counter goes on the token as it is
// created (CreateTokensThenForEffect's continuation names it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46b01055-4839-4b95-aa53-b954e9669e03",
		Name:         "Summoner's Sending",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(AtYourEndStep("Summoner's Sending — exile a creature card from a graveyard for a Spirit",
				summonersSendingExile),
				TargetCardInGraveyard("target creature card from a graveyard", Creature())),
				"Summoner's Sending — exile target creature card from a graveyard?"),
		},
	})
}

func summonersSendingExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	exiledID := ts[0].ID
	return ExileTarget{Target: exiledID, Then: func(ctx *Context, exiled bool) error {
		if !exiled {
			return nil
		}
		big := false
		if c, ok := ctx.Game.LookupCardForEffect(exiledID); ok {
			big = c.ManaValue() >= 4
		}
		bigN := 0
		if big {
			bigN = 1
		}
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(),
			Source:     ctx.Source(),
			Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("1/1 white Spirit with flying")), Count: 1}},
		}, ctx.TailN(summonersSendingCounter, bigN))
	}}.Apply(ctx)
}

// summonersSendingCounter is "if its mana value was 4 or greater, put a
// +1/+1 counter on the token" as a keyed tail (#4253 slice 6): whether it
// was is the tail's N. An on-disk key.
var summonersSendingCounter = CardTail("summoners-sending/counter-if-big", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N != 1 {
		return nil
	}
	for _, id := range o.IDs {
		if err := ctx.Game.AddCounterByForEffect(o.Tail.Player, id, game.CounterPlusOne, 1); err != nil {
			return err
		}
	}
	return nil
})
