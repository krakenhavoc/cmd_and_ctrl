package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nykthos Paragon — Enchantment Creature — Human Soldier {4}{W}{W},
// 4/6 (EDHREC rank 2952):
//
//	"Whenever you gain life, you may put that many +1/+1 counters on
//	 each creature you control. Do this only once each turn."
//
// The Theros lifegain payoff. b10YouGainedLife is the condition —
// one trigger per gain EVENT, lifelink included — and "that many" is
// the event's amount, captured as the trigger is built. The "you
// may" is the trigger's yes/no prompt; on yes the counters go on
// every creature the controller controls as the ability resolves
// (b28PutCountersOnEachCreatureYouControl — the Paragon included).
//
// "Do this only once each turn" limits the counters, not the trigger
// (b28DoneThisTurn): every gain triggers it and asks its "you may" as
// it resolves (CR 603.5), the first "yes" puts the counters and uses up
// the turn, and declining does not. Two gains at once are two triggers,
// so the controller chooses which one to accept. The gate is also read
// as the ability triggers, so no trigger is made once the turn's
// counters are down.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "367ac7e2-5056-48f3-a296-87d6cb1f7b54",
		Name:         "Nykthos Paragon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10YouGainedLife(ev, source) &&
					!b28DoneThisTurn(g, source.InstanceID, b28NykthosParagonDone)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Nykthos Paragon — put that many +1/+1 counters on each creature you control?"},
			Key:            b28NykthosParagonLabel,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if b28DoneThisTurn(g, item.SourceCardID, b28NykthosParagonDone) {
					return nil
				}
				b28NoteDoneThisTurn(g, item.SourceCardID, b28NykthosParagonDone)
				return b28PutCountersOnEachCreatureYouControl(g, item, item.Trigger.Event.Amount)
			},
		}},
	})
}

// b28NykthosParagonLabel is the trigger's stack label.
const b28NykthosParagonLabel = "Nykthos Paragon — that many +1/+1 counters on each creature you control"

// b28NykthosParagonDone is the once-each-turn tally's label: the
// counters were put this turn (b28DoneThisTurn).
const b28NykthosParagonDone = "Nykthos Paragon — counters put this turn"
