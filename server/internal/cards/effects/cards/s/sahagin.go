package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sahagin — Creature — Merfolk Warrior {1}{U}, 1/3 (Final Fantasy,
// #4462):
//
//	"Whenever you cast a noncreature spell, if at least four mana was
//	 spent to cast it, put a +1/+1 counter on this creature and it can't
//	 be blocked this turn."
//
// Blazing Bomb's cast trigger, read off the spell's payment record as it
// is cast (the amount never changes afterwards, so the intervening "if"
// reads the same at resolution). Both halves apply only while the
// creature is still the object the trigger came from (CR 400.7); a
// creature already blocked stays blocked (CR 509.1h).
//
// No simplification.
func init() {
	const label = "Sahagin — put a +1/+1 counter on this creature and it can't be blocked this turn"
	grow := effects.On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return effects.YouCast(effects.Noncreature())(ev, source, lki, g) && g.StackItemPaidForEffect(ev.CardID).Spent().Total() >= 4
	}, label, func(g *game.Game, item *game.StackItem) error {
		if !effects.SourceIsStillThisPermanent(g, item) {
			return nil
		}
		ctx := effects.NewContext(g, item)
		if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
			Label: "Sahagin — can't be blocked"}.Apply(ctx)
	})
	effects.Register(effects.Spec{
		OracleID:       "24310f49-2156-4c0d-b3b1-40ca3182474f",
		Name:           "Sahagin",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered:      []game.TriggeredAbility{grow},
	})
}
