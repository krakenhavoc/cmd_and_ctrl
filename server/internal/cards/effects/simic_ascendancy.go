package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "57c1a98e-c4fb-4144-9e3e-d0fe6789ba26",
		Name:         "Simic Ascendancy",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}{U}: Put a +1/+1 counter on target creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{G}{U}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  PlusOneCounterOnTheTarget,
		}},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := b39PlusOneCountersOnCreatureYouControl(ev, source, g)
					return ok
				},
				Key: "Simic Ascendancy — put that many growth counters on this enchantment",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					amount, _ := b39PlusOneCountersOnCreatureYouControl(ev, source, g)
					item := game.NewTriggeredItem(source, "Simic Ascendancy — put that many growth counters on this enchantment")
					item.Params.Amount = amount
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) || item.Params.Amount <= 0 {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: cmdr2B23GrowthCounter, N: item.Params.Amount}.Apply(NewContext(g, item))
				},
			},
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && source.Counters[cmdr2B23GrowthCounter] >= 20
			}, "Simic Ascendancy — if it has twenty or more growth counters, you win the game", cmdr2B23SimicAscendancyWin),
		},
	})
}

// cmdr2B23SimicAscendancyWin re-checks the intervening if on the same
// object (CR 603.4, CR 400.7) and wins the game.
func cmdr2B23SimicAscendancyWin(g *game.Game, item *game.StackItem) error {
	if item.Controller == uuid.Nil || sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Counters[cmdr2B23GrowthCounter] < 20 {
		return nil
	}
	return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
}
