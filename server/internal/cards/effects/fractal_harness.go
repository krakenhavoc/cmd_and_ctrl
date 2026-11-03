package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fractal Harness — Artifact — Equipment {X}{2}{G} (Secrets of
// Strixhaven Commander, #4834):
//
//	"When this Equipment enters, create a 0/0 green and blue Fractal
//	 creature token. Put X +1/+1 counters on it and attach this
//	 Equipment to it.
//	 Whenever equipped creature attacks, double the number of +1/+1
//	 counters on it.
//	 Equip {2}"
//
// X is the value announced for the Equipment spell (CR 107.3m), read off
// the permanent as the trigger is built and carried on the item, so a
// Harness put onto the battlefield without being cast, or cast for 0,
// makes a 0/0 that dies (the ruling, CR 704.5f). The token is created
// through the keyed door; the counters are put on it once it has landed,
// and the attach waits for the counters to settle (a CR 616 ordering
// prompt over the placement is a restore point). The Fractal is a
// described token, its colour words read through the card's text.
//
// The attack trigger doubles by putting as many +1/+1 counters on the
// creature as it has (the ruling), so a counter doubler applies. "It" is
// the creature that attacked; one that has left by resolution is left
// alone.
//
// No simplification.
func init() {
	const label = "Fractal Harness — create a Fractal with X +1/+1 counters and attach this to it"
	Register(Spec{
		OracleID:     "09342fd7-3dad-431e-8b87-de8a9287e757",
		Name:         "Fractal Harness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       label,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, label)
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: fractalHarnessCreateFractal,
			},
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Fractal Harness — double the +1/+1 counters on equipped creature", func(g *game.Game, item *game.StackItem) error {
				return b23DoublePlusOneCountersOn(NewContext(g, item), item.Trigger.Event.CardID)
			}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// fractalHarnessCreateFractal is the enters trigger's first sentence.
func fractalHarnessCreateFractal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(TokenCard("0/0 green and blue Fractal")),
			Count:    1,
		}},
	}, withN(ctx.ItemTail(fractalHarnessFractalCreated), item.Params.Amount))
}

// fractalHarnessFractalCreated puts the X counters on the Fractal that
// arrived; X rides the tail as its N. An on-disk key.
var fractalHarnessFractalCreated = CardTail("fractal-harness/fractal-created", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	then := ctx.ItemTail(fractalHarnessAttach)
	then.Carry = []uuid.UUID{o.IDs[0]}
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, o.IDs[0], game.CounterPlusOne, o.Tail.N, then)
})

// fractalHarnessAttach attaches the Harness to the Fractal once its
// counters have settled. An on-disk key.
var fractalHarnessAttach = CardTail("fractal-harness/attach", func(ctx *Context, o game.TailOutcome) error {
	fractal := carryFirst(o.Tail)
	if fractal == uuid.Nil {
		return nil
	}
	return ctx.Game.AttachSourceForEffect(ctx.Item, game.TargetRef{Kind: game.TargetCard, ID: fractal})
})
