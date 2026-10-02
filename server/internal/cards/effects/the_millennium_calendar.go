package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Millennium Calendar — Legendary Artifact {1}:
//
//	"Whenever you untap one or more permanents during your untap step,
//	 put that many time counters on The Millennium Calendar.
//	 {2}, {T}: Double the number of time counters on The Millennium
//	 Calendar.
//	 When there are 1,000 or more time counters on The Millennium
//	 Calendar, sacrifice it and each opponent loses 1,000 life."
//
// ADR 0107 §1 (#1858):
//
//   - The untap trigger is "one or more", so it fires once for the CR
//     502.3 untap of the controller's own untap step (one event batch),
//     and "that many" is counted on resolution from the untap events of
//     that batch whose permanent the controller controlled. Permanents a
//     Seedborn Muse untaps in another player's untap step are not "your
//     untap step", and a permanent kept tapped is not untapped.
//   - Doubling is CR 701.10e: put as many time counters on it as it
//     already has, so a counter doubler applies to the addition.
//   - The finale is a CR 603.8 state trigger. Both halves happen whether
//     or not the sacrifice does — there is no "if you do".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f250443-8d5e-46c9-920e-8e9373780e32",
		Name:         "The Millennium Calendar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventUntapCard, millenniumYouUntapInYourUntapStep,
				"The Millennium Calendar — put that many time counters", millenniumTimeCounters)),
			WhenThisHasAtLeast(game.CounterTime, 1000, "The Millennium Calendar — sacrifice it and each opponent loses 1,000 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := SacrificeThisIfStillOnBattlefield(g, item); err != nil {
						return err
					}
					return eachOpponentLosesLife(g, item, 1000)
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Double the number of time counters on The Millennium Calendar.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !onBattlefield(g, item.SourceCardID) || c.Counters[game.CounterTime] <= 0 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: c.Counters[game.CounterTime]}.Apply(NewContext(g, item))
			},
		}},
	})
}

// millenniumYouUntapInYourUntapStep is "you untap a permanent during your
// untap step": an untap of a permanent the source's controller controls,
// in the untap step of that player's own turn.
func millenniumYouUntapInYourUntapStep(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller || g.Turn.Step != game.StepUntap {
		return false
	}
	seat := g.Turn.ActiveSeat
	return seat >= 0 && seat < len(g.Seats) && g.Seats[seat].ID == source.Controller
}

// millenniumTimeCounters puts one time counter on the Calendar for each
// permanent its controller untapped in the batch that triggered it.
func millenniumTimeCounters(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	batch := item.Trigger.Event.Batch
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Batch == batch && ev.Kind == game.EventUntapCard && ev.Actor == item.Controller {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: n}.Apply(NewContext(g, item))
}
