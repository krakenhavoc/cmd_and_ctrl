package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Idol of False Gods — Kindred Artifact — Eldrazi {2}:
//
//	"{1}{C}, {T}: Create a 0/1 colorless Eldrazi Spawn creature token
//	 with "Sacrifice this token: Add {C}."
//	 Whenever another Eldrazi you control dies, put a +1/+1 counter on
//	 this artifact.
//	 As long as this artifact has eight or more +1/+1 counters on it,
//	 it's a 0/0 creature in addition to its other types and it has
//	 annihilator 2."
//
// The Spawn is the shared token. "Another Eldrazi you control dies" is
// any Eldrazi permanent, creature or not (the 2024-06-07 ruling), read
// as it last existed on the battlefield (CR 603.10a). At eight or more
// counters three statics switch on together: the creature type is
// added in layer 4, its base power and toughness are 0/0 in layer 7b
// (so the counters make it an 8/8 or bigger), and it has annihilator 2
// in layer 6 — the keyword the engine turns into an attack trigger
// (ADR 0113 §2).
//
// The three statics are ONE printed effect that starts in layer 4. An
// Idol that loses its abilities stays a creature (layer 4 comes before a
// layer 6 removal, CR 613.6, ADR 0067 section 2), and its layer 6 and 7b
// halves declare ContinuesAfterRemoval, so they carry on past the
// removal in timestamp order (#3862): the 0/0 base keeps applying ahead
// of an older effect's base power and toughness, and the annihilator
// survives a removal older than the effect but not one newer than it.
//
// "Older than the effect" is the ruling's measure, not the Idol's
// entry: "If Idol of False Gods loses its abilities before gaining its
// eighth +1/+1 counter, it will have annihilator 2; otherwise, it
// won't." The effect begins at the eighth counter, so all three statics
// declare StartsAtCounters and sort at that moment
// (game/counter_thresholds.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b30d583b-e770-47e8-9f26-fca7ef43285c",
		Name:         "Idol of False Gods",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{C}, {T}: Create a 0/1 colorless Eldrazi Spawn creature token with \"Sacrifice this token: Add {C}.\"",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{C}"), TapCost()),
			Effect:  Do(CreateToken{Template: EldraziSpawnToken(), N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, anotherEldraziYouControlDied,
				"Idol of False Gods — put a +1/+1 counter on this artifact", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Static: idolOfFalseGodsStatics(),
	})
}

// anotherEldraziYouControlDied is "Whenever another Eldrazi you control
// dies": an Eldrazi permanent, creature or not, the source's controller
// controlled, put into a graveyard from the battlefield.
func anotherEldraziYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftUnderControlOf(ev, dead) == source.Controller && leftAsSubtype(ev, dead, "Eldrazi")
}

// idolHasEightCounters is the statics' condition.
func idolHasEightCounters(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && target.Counters[game.CounterPlusOne] >= 8
}

// idolEighthCounter is the moment the statics' effect begins.
var idolEighthCounter = game.CounterThreshold{Kind: game.CounterPlusOne, AtLeast: 8}

// idolOfFalseGodsStatics are the three statics of "it's a 0/0 creature
// in addition to its other types and it has annihilator 2".
func idolOfFalseGodsStatics() []game.StaticAbility {
	return []game.StaticAbility{
		{
			// The one row the three statics show (the other two continue
			// it), so it carries the whole sentence.
			Label:            "As long as this artifact has eight or more +1/+1 counters on it, it's a 0/0 creature in addition to its other types and it has annihilator 2.",
			Layer:            game.Layer4Type,
			AppliesTo:        idolHasEightCounters,
			StartsAtCounters: idolEighthCounter,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Types, "Creature") {
					c.Types = append(c.Types, "Creature")
				}
			},
		},
		{
			Label:     "As long as this artifact has eight or more +1/+1 counters on it, it's a 0/0 creature.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: idolHasEightCounters,
			// One printed effect that starts in layer 4: CR 613.6 keeps
			// its later layers applying past an older ability removal.
			ContinuesAfterRemoval: true,
			StartsAtCounters:      idolEighthCounter,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power, c.Toughness = 0, 0
			},
		},
		{
			Label:     "As long as this artifact has eight or more +1/+1 counters on it, it has annihilator 2.",
			Layer:     game.Layer6Ability,
			AppliesTo: idolHasEightCounters,
			// The same effect, so CR 613.6 carries it past a removal that
			// silences the Idol in this very layer, in timestamp order.
			ContinuesAfterRemoval: true,
			StartsAtCounters:      idolEighthCounter,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "annihilator 2")
			},
		},
	}
}
