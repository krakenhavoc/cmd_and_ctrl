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
// Sandbox simplification, declared: an Idol that loses its abilities
// loses all three statics, where the rules keep the type change (layer
// 4 applies before the removal in layer 6, the 2024-06-07 ruling). The
// engine has no dependency ordering between one permanent's own
// statics and a removal.
func init() {
	Register(Spec{
		OracleID:     "b30d583b-e770-47e8-9f26-fca7ef43285c",
		Name:         "Idol of False Gods",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"An effect that takes away Idol of False Gods's abilities also stops it being a creature."},
		Activated: []ActivatedAbility{{
			Label:   "{1}{C}, {T}: Create a 0/1 colorless Eldrazi Spawn creature token with \"Sacrifice this token: Add {C}.\"",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{C}"), TapCost()),
			Effect:  Do(CreateToken{Template: EldraziSpawnToken(), N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, anotherEldraziYouControlDied,
				"Idol of False Gods — put a +1/+1 counter on this artifact", b35PutCounterOnSelf),
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

// idolOfFalseGodsStatics are the three statics of "it's a 0/0 creature
// in addition to its other types and it has annihilator 2".
func idolOfFalseGodsStatics() []game.StaticAbility {
	return []game.StaticAbility{
		{
			Layer:     game.Layer4Type,
			AppliesTo: idolHasEightCounters,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Types, "Creature") {
					c.Types = append(c.Types, "Creature")
				}
			},
		},
		{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: idolHasEightCounters,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power, c.Toughness = 0, 0
			},
		},
		{
			Layer:     game.Layer6Ability,
			AppliesTo: idolHasEightCounters,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "annihilator 2")
			},
		},
	}
}
