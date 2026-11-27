package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sporogenesis — Enchantment {3}{G} (Urza's Saga, S82 OLD batch 732, #5624):
//
//	"At the beginning of your upkeep, you may put a fungus counter on target
//	 nontoken creature.
//	 Whenever a creature with a fungus counter on it dies, create a 1/1 green
//	 Saproling creature token for each fungus counter on that creature.
//	 When this enchantment leaves the battlefield, remove all fungus counters
//	 from all creatures."
//
// The upkeep trigger targets (any player's nontoken creature) and asks its "may"
// as it resolves (CR 603.5), the target rechecked first. The dies trigger is
// any creature that last had a fungus counter (CR 603.10a), and the number of
// Saprolings, all its controller's, is the number of counters it last had.
// Leaving the battlefield removes every fungus counter from every creature; a
// copy of Sporogenesis on the table does not stop the other from doing the same.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33005c63-fe9f-4051-b781-eacf1ca5eec6",
		Name:         "Sporogenesis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep("Sporogenesis — you may put a fungus counter on target nontoken creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.MayChoice{
						Question: "Sporogenesis — put a fungus counter on " + effects.CardNameFor(g, id) + "?",
						Then:     sporogenesisCounter,
						Carry:    []uuid.UUID{id},
					}.Apply(ctx)
				}),
				effects.TargetCreature("target nontoken creature", effects.Not(effects.IsTokenPredicate()))),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if _, died := effects.DiedCreature(ev, g); !died {
					return false
				}
				return g.LastKnownCountersForEffect(ev.CardID)[sporogenesisCounterKind] > 0
			}, "Sporogenesis — create a 1/1 green Saproling for each fungus counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					// The counters it last had, from the permanent's own
					// record (CR 608.2h): the engine's trigger-time snapshot
					// is spent once the trigger is on the stack.
					info, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					n := info.Counters[sporogenesisCounterKind]
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
				}),
			effects.WhenThisLeaves("Sporogenesis — remove all fungus counters from all creatures",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, c := range g.BattlefieldCardsForEffect() {
						if n := c.Counters[sporogenesisCounterKind]; n > 0 && c.IsCreature() {
							if err := (effects.AddCounter{Target: c.InstanceID, Kind: sporogenesisCounterKind, N: -n}).Apply(ctx.AsGroupMember()); err != nil {
								return err
							}
						}
					}
					return nil
				}),
		},
	})
}

// sporogenesisCounterKind is the counter Sporogenesis names.
const sporogenesisCounterKind = "fungus"

// sporogenesisCounter is the "yes": a fungus counter on the target, if it is
// still on the battlefield. An on-disk key.
var sporogenesisCounter = effects.MayYes("confirm/sporogenesis-fungus-counter", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	id := effects.CarryAt(r, 0)
	if id == uuid.Nil || !effects.OnBattlefield(ctx.Game, id) {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: sporogenesisCounterKind, N: 1}.Apply(ctx)
})
