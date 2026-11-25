package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarulf, Realm Eater — Legendary Creature — Wolf {1}{B}{G}, 3/3 (Secret Lair
// Drop):
//
//	"Whenever a permanent an opponent controls is put into a graveyard from
//	 the battlefield, put a +1/+1 counter on Sarulf.
//	 At the beginning of your upkeep, if Sarulf has one or more +1/+1
//	 counters on it, you may remove all of them. If you do, exile each other
//	 nonland permanent with mana value less than or equal to the number of
//	 counters removed this way."
//
// The first trigger reads who controlled the permanent as it last existed
// (CR 603.10a, LeftUnderControlOf), so a stolen creature that dies counts for
// its thief's opponents. The second has an intervening "if" (CR 603.4): it is
// checked when the upkeep begins and again as the ability resolves. Removing
// the counters is the "if you do"; the exile counts the ones actually removed.
//
// No simplification.
func init() {
	upkeep := effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source.Counters[game.CounterPlusOne] > 0
	}), "Sarulf, Realm Eater — you may remove all +1/+1 counters; if you do, exile each other nonland permanent with that mana value or less", sarulfUpkeep)
	upkeep.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Sarulf, Realm Eater: remove all +1/+1 counters and exile each other nonland permanent with mana value up to that many?"}
	effects.Register(effects.Spec{
		OracleID:     "789964f5-79c5-4329-b68f-b15b0d54b0b2",
		Name:         "Sarulf, Realm Eater",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.NewZone != game.ZoneGraveyard {
					return false
				}
				dead, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftUnderControlOf(ev, dead) != source.Controller
			}, "Sarulf, Realm Eater — +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.TriggerWithPurpose(upkeep, game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile, Partial: true}}),
		},
	})
}

func sarulfUpkeep(g *game.Game, item *game.StackItem) error {
	if !effects.B09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	source := item.SourceCardID
	n := effects.CounterCountOn(g, source, game.CounterPlusOne)
	if n <= 0 {
		return nil
	}
	if err := (effects.AddCounter{Target: source, Kind: game.CounterPlusOne, N: -n}).Apply(ctx); err != nil {
		return err
	}
	return effects.ExileAllMatching{Match: effects.And(effects.Not(effects.Land()), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.InstanceID != source
	}, effects.ManaValueLE(n))}.Apply(ctx)
}
