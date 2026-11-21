package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sahir, Visitor in Darkness — Legendary Creature — Human Wizard {1}{B}{B},
// 3/3 (Mystery Booster Commander, #5948):
//
//	"Eminence — At the beginning of your upkeep, Sahir deals 1 damage to you.
//	 This ability triggers only if Sahir is in the command zone or on the
//	 battlefield.
//	 Whenever a source you control deals exactly 1 damage to you and whenever
//	 a spell or ability you control causes you to lose exactly 1 life, put a
//	 +1/+1 counter on Sahir. If this is the second time this ability has
//	 resolved this turn, draw a card."
//
// The eminence trigger is Edgar Markov's zone pair (EminenceTrigger); the
// damage comes from Sahir wherever he is. The counter ability is one trigger
// with two conditions (OnAny), as printed:
//
//   - an EventDealDamage of exactly 1 to you from a source you control
//     (a creature, a spell on the stack, Sahir himself, wherever it is), and
//   - an EventChangeLife of exactly -1 for you whose source is a card you
//     control, which is life an effect made you lose. Life paid as a cost
//     (Event.LifePayment), rad-counter loss and damage are not it (CR 119.3,
//     119.4); damage already has its own half.
//
// It is a battlefield ability only. "The second time this ability has
// resolved this turn" is Jake Sisko's per-object tally (ResolvedThisTurn),
// counting the resolution in progress, so only the second draws, not the
// third. The counter goes on Sahir only while he is the same object on the
// battlefield (CR 400.7); the draw is not conditional on that, as printed.
//
// No simplification.
func init() {
	const label = "Sahir, Visitor in Darkness — put a +1/+1 counter on Sahir; if this is the second time this ability has resolved this turn, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "e1a8f5cf-9b59-46c6-8e41-6e44225ec0f4",
		Name:         "Sahir, Visitor in Darkness",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.EminenceTrigger(effects.AtYourUpkeep("Sahir, Visitor in Darkness — eminence: Sahir deals 1 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
				})),
			effects.OnAny([]game.EventKind{game.EventDealDamage, game.EventChangeLife}, sahirExactlyOne, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.OnBattlefield(g, item.SourceCardID) && !ctx.IsNewSourceObject(item.SourceCardID) {
						if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					if g.ResolvedThisTurn(item.SourceCardID, label) != 2 {
						return nil
					}
					return effects.DrawCards{N: 1}.Apply(ctx)
				}),
		},
	})
}

// sahirExactlyOne is both halves of the counter ability's condition.
func sahirExactlyOne(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	you := source.Controller
	if ev.Target != you {
		return false
	}
	switch ev.Kind {
	case game.EventDealDamage:
		return ev.Amount == 1 && sahirControls(g, ev.Source, you)
	case game.EventChangeLife:
		return ev.Amount == -1 && !ev.FromRadiation && !ev.LifePayment && sahirControls(g, ev.Source, you)
	}
	return false
}

// sahirControls is "a source you control": a card of yours, wherever it is.
func sahirControls(g *game.Game, id uuid.UUID, you uuid.UUID) bool {
	if id == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Controller == you
}
