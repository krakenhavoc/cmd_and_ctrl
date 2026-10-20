package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gatewatch Beacon — Artifact {2}{W}:
//
//	"This artifact enters with three loyalty counters on it.
//	 {T}: Add {W}.
//	 Whenever a planeswalker you control enters, if this artifact has
//	 loyalty counters on it, you may move a loyalty counter from this
//	 artifact onto that planeswalker."
//
// The three counters are a CR 614.1c self-replacement on the entry
// (b10EntersWithCounters), so they are on the artifact before anything
// sees it. The Beacon is not a planeswalker, so its loyalty counters are
// only counters: CR 704.5i never looks at it, and nothing stamps a
// starting loyalty on it.
//
// The trigger has an intervening "if" (CR 603.4): it triggers only while
// the Beacon has a loyalty counter, and it checks again before its "you
// may" is asked (InterveningIf, ADR 0197), so a Beacon emptied in
// response asks nothing. "That planeswalker" is the object that entered:
// one that left the battlefield in response is gone (CR 400.7) and gets
// nothing, and the Beacon keeps its counter. The move is a removal from
// the Beacon and a placement on the walker, so a counter replacement
// (Doubling Season) sees the placement.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "2a31cd16-fb30-4bb0-b775-d5e8b9221dd4",
		Name:         "Gatewatch Beacon",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterLoyalty, 3, "Gatewatch Beacon — enters with three loyalty counters"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(gatewatchBeaconHasLoyalty{}, Optional(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					entered, ok := enteredUnderYourControl(ev, source, g, true)
					return ok && entered.IsPlaneswalker() && source.Counters[game.CounterLoyalty] > 0
				}, "Gatewatch Beacon — move a loyalty counter onto that planeswalker", gatewatchBeaconMove),
				"Move a loyalty counter from Gatewatch Beacon onto that planeswalker?")),
		},
	})
}

// gatewatchBeaconHasLoyalty is the intervening "if this artifact has
// loyalty counters on it", read off the Beacon as it is now. A Beacon
// that has left the battlefield has none.
type gatewatchBeaconHasLoyalty struct{}

func (gatewatchBeaconHasLoyalty) Holds(ctx *Context) bool {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return false
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
	return ok && c.Counters[game.CounterLoyalty] > 0
}

func (gatewatchBeaconHasLoyalty) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// gatewatchBeaconMove moves one loyalty counter from the Beacon onto the
// planeswalker whose entry triggered it, if that walker is still on the
// battlefield as the same object.
func gatewatchBeaconMove(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	walker := item.Trigger.Event.CardID
	if !onBattlefield(g, walker) {
		return nil
	}
	if !(gatewatchBeaconHasLoyalty{}).Holds(ctx) {
		return nil
	}
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterLoyalty, N: -1}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: walker, Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
}
