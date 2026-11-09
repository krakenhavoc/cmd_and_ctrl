package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyknight Squire — Creature — Cat Scout {1}{W}, 1/1 (Foundations,
// #4465):
//
//	"Whenever another creature you control enters, put a +1/+1 counter
//	 on this creature.
//	 As long as this creature has three or more +1/+1 counters on it,
//	 it has flying and is a Knight in addition to its other types."
//
// The trigger is a one-sentence counter on itself (Do(CounterOnThis)), one
// trigger per creature that enters, so a creature entering with others
// triggers once for each (the ruling). The second ability is two statics
// on the Squire alone, both on while it has three or more +1/+1 counters:
// the Knight subtype in layer 4 (CR 613.1d) and flying in layer 6 (CR
// 613.1f). Each declares StartsAtCounters, so the effect is timestamped at
// the third counter (CR 613.7) and an ability removal that arrived earlier
// applies first. Flying gained after blockers are declared does not undo
// a block (CR 509.1b), which the engine's combat already honours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e497f6b-69e1-4fab-bb04-f687c99a8310",
		Name:         "Skyknight Squire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Skyknight Squire — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Static: []game.StaticAbility{
			{
				Label:            "As long as this creature has three or more +1/+1 counters on it, it is a Knight in addition to its other types.",
				Layer:            game.Layer4Type,
				AppliesTo:        skyknightSquireHasThreeCounters,
				StartsAtCounters: skyknightSquireThirdCounter,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !hasFold(c.Subtypes, "Knight") {
						c.Subtypes = append(c.Subtypes, "Knight")
					}
				},
			},
			{
				Label:            "As long as this creature has three or more +1/+1 counters on it, it has flying.",
				Layer:            game.Layer6Ability,
				AppliesTo:        skyknightSquireHasThreeCounters,
				StartsAtCounters: skyknightSquireThirdCounter,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
				},
			},
		},
	})
}

// skyknightSquireThirdCounter is the moment the statics' effect begins.
var skyknightSquireThirdCounter = game.CounterThreshold{Kind: game.CounterPlusOne, AtLeast: 3}

// skyknightSquireHasThreeCounters is the statics' condition: the Squire
// itself, with three or more +1/+1 counters.
func skyknightSquireHasThreeCounters(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && target.Counters[game.CounterPlusOne] >= 3
}
