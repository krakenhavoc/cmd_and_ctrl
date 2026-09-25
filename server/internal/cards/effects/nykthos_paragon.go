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
// "Do this only once each turn" is two checks in AppliesTo: an
// ability with this label already put on the stack this turn
// (b11TriggeredThisTurn — the harvester logs EventTrigger only when
// the item is queued, so a declined prompt does not use up the turn,
// exactly as declining the printed "you may" does not) and a prompt
// from the Paragon still open (b28TriggerPromptPendingFrom — two
// lifelinkers connecting in one combat are two gain events in one
// mutation, and without the dedup both would ask and both could be
// accepted).
//
// Sandbox simplification, declared: with several gains in one
// mutation only the first is offered, where the printed card lets
// the controller choose which gain's trigger to accept. Weaker
// (never two placements), never stronger.
func init() {
	Register(Spec{
		OracleID:     "367ac7e2-5056-48f3-a296-87d6cb1f7b54",
		Name:         "Nykthos Paragon",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"When you gain life several times at once, only the first gain offers the counters."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10YouGainedLife(ev, source) &&
					!b11TriggeredThisTurn(g, source.InstanceID, b28NykthosParagonLabel) &&
					!b28TriggerPromptPendingFrom(g, source)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Nykthos Paragon — put that many +1/+1 counters on each creature you control?"},
			Key:            b28NykthosParagonLabel,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b28PutCountersOnEachCreatureYouControl(g, item, item.Trigger.Event.Amount)
			},
		}},
	})
}

// b28NykthosParagonLabel is the trigger's stack label — named because
// the once-per-turn tally matches on it.
const b28NykthosParagonLabel = "Nykthos Paragon — that many +1/+1 counters on each creature you control"
