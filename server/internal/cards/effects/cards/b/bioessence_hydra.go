package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bioessence Hydra — Creature — Hydra Mutant {3}{G}{U}, 4/4 (War of the Spark):
//
//	"Trample
//	 This creature enters with a +1/+1 counter on it for each loyalty counter on
//	 planeswalkers you control.
//	 Whenever one or more loyalty counters are put on planeswalkers you control,
//	 put that many +1/+1 counters on this creature."
//
// The entry counters are a replacement (CR 614.1c) counted as the Hydra enters,
// over the loyalty counters then on the planeswalkers its controller has. The
// trigger reads the delta of one placement (the event carries the new total),
// and puts that many counters on the Hydra while it is still the same permanent.
// A planeswalker entering with its own loyalty is not a placement that triggers
// it (CR 122.6 counters from entering); a loyalty ability's cost or effect, or
// an effect that adds loyalty, is.
//
// No simplification.
const bioessenceHydraLabel = "Bioessence Hydra — put that many +1/+1 counters on this creature"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "1f79de2a-6cef-4285-b008-5b5527c406ab",
		Name:            "Bioessence Hydra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == src.Controller && c.IsPlaneswalker() {
						n += c.Counters[game.CounterLoyalty]
					}
				}
				return n
			}, "Bioessence Hydra: enters with a +1/+1 counter for each loyalty counter on planeswalkers you control"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return oldB167LoyaltyPlacedOnYourPlaneswalker(ev, source.Controller, g) > 0
			},
			Key: bioessenceHydraLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, bioessenceHydraLabel)
				item.Params.Amount = oldB167LoyaltyPlacedOnYourPlaneswalker(ev, source.Controller, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// oldB167LoyaltyPlacedOnYourPlaneswalker is how many loyalty counters an
// EventCounterPlaced put on a planeswalker controller controls, or zero.
func oldB167LoyaltyPlacedOnYourPlaneswalker(ev game.Event, controller uuid.UUID, g *game.Game) int {
	if ev.Kind != game.EventCounterPlaced || ev.Label != game.CounterLoyalty || ev.Target == uuid.Nil {
		return 0
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !c.IsPlaneswalker() || c.Controller != controller || !g.Battlefield.Contains(ev.Target) {
		return 0
	}
	return effects.B33CountersPlacedDelta(ev, game.CounterLoyalty, g)
}
