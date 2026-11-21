package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edgar Markov — Legendary Creature — Vampire Knight {3}{R}{W}{B}, 4/4
// (EDHREC rank 3158):
//
//	"Eminence — Whenever you cast another Vampire spell, if Edgar is in
//	 the command zone or on the battlefield, create a 1/1 black Vampire
//	 creature token.
//	 First strike, haste
//	 Whenever Edgar attacks, put a +1/+1 counter on each Vampire you
//	 control."
//
// The eminence trigger works from the command zone (#2802): Edgar never
// has to be cast for every Vampire its owner casts to bring a token.
// "Another" is the cast spell not being Edgar itself; Edgar cast from
// the command zone has already left it as it is cast, so it never sees
// its own cast anyway.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "41e2790d-49f5-4e98-b8d9-04179f47f13a",
		Name:            "Edgar Markov",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.EminenceTrigger(effects.On(game.EventCast, youCastAnotherVampire,
				"Edgar Markov — create a 1/1 Vampire",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Vampire"), N: 1}))),
			effects.WheneverThisAttacks("Edgar Markov — a +1/+1 counter on each Vampire you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.PermanentsControlledByMatching(g, item.Controller, effects.Subtype("Vampire")) {
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

// youCastAnotherVampire is "Whenever you cast another Vampire spell".
func youCastAnotherVampire(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ev.CardID != source.InstanceID && effects.YouCast(effects.Subtype("Vampire"))(ev, source, lki, g)
}
