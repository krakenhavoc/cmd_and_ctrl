package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Simic Ascendancy — Enchantment {G}{U} (Bloomburrow Commander, #5368):
//
//	"{1}{G}{U}: Put a +1/+1 counter on target creature you control.
//	 Whenever one or more +1/+1 counters are put on a creature you control,
//	 put that many growth counters on this enchantment.
//	 At the beginning of your upkeep, if this enchantment has twenty or more
//	 growth counters on it, you win the game."
//
// The counter trigger is Shalai and Hallar's: one per creature touched,
// whoever puts them and including the counters a creature enters with, with
// the batch's size as the count. The win is an intervening "if" (CR 603.4)
// read as the upkeep begins and again as it resolves, on the same object
// (Chance Encounter's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57c1a98e-c4fb-4144-9e3e-d0fe6789ba26",
		Name:         "Simic Ascendancy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{U}: Put a +1/+1 counter on target creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{G}{U}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.PlusOneCounterOnTheTarget,
		}},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := effects.B39PlusOneCountersOnCreatureYouControl(ev, source, g)
					return ok
				},
				Key: "Simic Ascendancy — put that many growth counters on this enchantment",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					amount, _ := effects.B39PlusOneCountersOnCreatureYouControl(ev, source, g)
					item := game.NewTriggeredItem(source, "Simic Ascendancy — put that many growth counters on this enchantment")
					item.Params.Amount = amount
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || item.Params.Amount <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: effects.Cmdr2B23GrowthCounter, N: item.Params.Amount}.Apply(effects.NewContext(g, item))
				},
			},
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && source.Counters[effects.Cmdr2B23GrowthCounter] >= 20
			}, "Simic Ascendancy — if it has twenty or more growth counters, you win the game", cmdr2B23SimicAscendancyWin),
		},
	})
}

// cmdr2B23SimicAscendancyWin re-checks the intervening if on the same
// object (CR 603.4, CR 400.7) and wins the game.
func cmdr2B23SimicAscendancyWin(g *game.Game, item *game.StackItem) error {
	if item.Controller == uuid.Nil || effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Counters[effects.Cmdr2B23GrowthCounter] < 20 {
		return nil
	}
	return effects.WinTheGame{Player: item.Controller}.Apply(effects.NewContext(g, item))
}
