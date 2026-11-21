package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Donatello, Mutant Mechanic — Legendary Creature — Mutant Ninja Turtle
// {3}{U}, 3/5 (Teenage Mutant Ninja Turtles, #4457):
//
//	"{T}: Put three +1/+1 counters on target artifact you control. If it
//	 isn't a creature, it becomes a 0/0 Robot creature in addition to its
//	 other types. Activate only as a sorcery.
//	 Whenever an artifact you control is put into a graveyard from the
//	 battlefield, if it had counters on it, put those counters on up to
//	 one target artifact or creature you control."
//
// The activated ability is Tezzeret, Cruel Captain's emblem body
// (tmtB8RobotCountersOnFirstTarget): an indefinite type change for a
// noncreature artifact, then the counters.
//
// The trigger is Resourceful Defense's "those counters": it watches an
// artifact you controlled, as it last existed (CR 603.10a), going to a
// graveyard from the battlefield, tokens included. "If it had counters
// on it" is an intervening if read from the engine's last-known
// counters, and the counters put are every kind it had, in the numbers
// it had, read again at resolution from the same record. The target is
// "up to one", so the trigger is never removed for want of one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6a7838d-380f-4752-a647-c55c2a0eb6e2",
		Name:         "Donatello, Mutant Mechanic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Put three +1/+1 counters on target artifact you control. If it isn't a creature, it becomes a 0/0 Robot creature in addition to its other types. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl()),
			Effect:       effects.TmtB8RobotCountersOnFirstTarget("Donatello, Mutant Mechanic — becomes a 0/0 Robot creature"),
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B23ArtifactPutIntoGraveyardFromBattlefield(ev, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftUnderControlOf(ev, c) == source.Controller &&
					len(g.LastKnownCountersForEffect(ev.CardID)) > 0
			},
			Targets: effects.TargetPermanent("up to one target artifact or creature you control",
				effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()).WithCount(0, 1),
			Key:    "Donatello, Mutant Mechanic — put those counters on up to one target artifact or creature you control",
			Effect: effects.ResourcefulDefensePutThoseCounters,
		}},
	})
}
