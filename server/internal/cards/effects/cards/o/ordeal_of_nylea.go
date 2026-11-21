package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ordeal of Nylea — Enchantment — Aura {1}{G} (Foundations, #4465):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, put a +1/+1 counter on it. Then
//	 if it has three or more +1/+1 counters on it, sacrifice this Aura.
//	 When you sacrifice this Aura, search your library for up to two basic
//	 land cards, put them onto the battlefield tapped, then shuffle."
//
// Two abilities. The attack trigger puts the counter and then, as part of
// the same resolution (the first ruling), counts the creature's +1/+1
// counters as it is now: a third counter put on any other way waits for
// the next attack. The sacrifice trigger watches EventSacrifice, which
// fires while the Aura is still on the battlefield, for THIS Aura
// sacrificed by its controller, so it triggers for any sacrifice (the
// second ruling), not only its own: a Aura that falls off and goes to the
// graveyard as a state-based action is not sacrificed and doesn't. Each
// search is an "up to two" with a stated quality, so it may find fewer.
//
// No simplification.
func init() {
	const name = "Ordeal of Nylea"
	effects.Register(effects.Spec{
		OracleID:     "5fff2bf4-9e26-4815-9bf6-672ed7e55b96",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks(name+" — put a +1/+1 counter on it, then sacrifice this Aura if it has three or more",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					host := item.Trigger.Event.CardID
					ctx := effects.NewContext(g, item)
					if err := (effects.AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					c, ok := g.LookupCardForEffect(host)
					if !ok || c.Counters[game.CounterPlusOne] < 3 {
						return nil
					}
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
				}),
			{
				Watches: []game.EventKind{game.EventSacrifice},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID && ev.Actor == source.Controller
				},
				Key:     name + " — search for up to two basic land cards, put them onto the battlefield tapped",
				Purpose: game.Purpose{Lands: 2},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         2,
						TappedOnEntry: true,
						Shuffle:       true,
						Reason:        name + " — search your library for up to two basic land cards",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
