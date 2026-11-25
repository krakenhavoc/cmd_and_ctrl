package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yorvo, Lord of Garenbrig — Legendary Creature — Giant Noble {G}{G}{G},
// 0/0 (Throne of Eldraine, #5516):
//
//	"Yorvo enters with four +1/+1 counters on it.
//	 Whenever another green creature you control enters, put a +1/+1
//	 counter on Yorvo. Then if that creature's power is greater than
//	 Yorvo's power, put another +1/+1 counter on Yorvo."
//
// The first counter goes on, then the two powers are compared (CR 608.2c):
// Yorvo's live power, and the entering creature's power as it is as the
// ability resolves, or its last-known power once it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00a15abc-bdac-4977-90a0-f08d41941d01",
		Name:         "Yorvo, Lord of Garenbrig",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 4, "Yorvo, Lord of Garenbrig: enters with four +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, s82OldB144AnotherGreenCreatureEntered,
				"Yorvo, Lord of Garenbrig — put a +1/+1 counter on Yorvo, then another if that creature's power is greater",
				s82OldB144YorvoGrows),
		},
	})
}

func s82OldB144AnotherGreenCreatureEntered(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !effects.AnotherCreatureEnteredUnderYourControl(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasColor(source.TextColor("G"))
}

func s82OldB144YorvoGrows(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := ctx.Source()
	if !effects.OnBattlefield(g, src) {
		return nil
	}
	entered, haveEntered := ctx.TriggeringPermanent()
	if err := (effects.AddCounter{Target: src, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if !haveEntered {
		return nil
	}
	yorvo, ok := g.LookupCardForEffect(src)
	if !ok || entered.Power <= yorvo.CurrentPower() {
		return nil
	}
	return effects.AddCounter{Target: src, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
