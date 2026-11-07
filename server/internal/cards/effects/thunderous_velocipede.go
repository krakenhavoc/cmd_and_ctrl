package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thunderous Velocipede — Artifact — Vehicle {1}{G}{G}, 5/5 (Aetherdrift,
// #4464):
//
//	"Trample
//	 Each other Vehicle and creature you control enters with an
//	 additional +1/+1 counter on it if its mana value is 4 or less.
//	 Otherwise, it enters with three additional +1/+1 counters on it.
//	 Crew 3"
//
// Renata, Called to the Hunt's CR 614.1c replacement on the entry of a
// permanent under the Velocipede's controller (tokens included), with
// "Vehicle or creature" as the test and the count read off the entering
// object's mana value: one counter at 4 or less, three above. A token
// that copies nothing has mana value 0 (CR 202.3), so it gets one. The
// Velocipede itself is excluded. Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "899007b5-6a07-4491-9104-bb537106b07d",
		Name:            "Thunderous Velocipede",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    []game.ReplacementEffect{thunderousVelocipedeEntersWithCounters()},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Thunderous Velocipede"),
		}},
	})
}

// thunderousVelocipedeEntersWithCounters is the Velocipede's second
// paragraph: another Vehicle or creature entering under its controller's
// control gets one extra +1/+1 counter at mana value 4 or less, three
// otherwise.
func thunderousVelocipedeEntersWithCounters() game.ReplacementEffect {
	entering := func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) (game.Card, bool) {
		if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
			return game.Card{}, false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || c.Controller != src.Controller || (!c.IsCreature() && !c.HasSubtype("Vehicle")) {
			return game.Card{}, false
		}
		return c, true
	}
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			_, ok := entering(ev, g, src)
			return ok
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			c, ok := entering(ev, g, src)
			if !ok {
				return nil
			}
			n := 1
			if c.ManaValue() > 4 {
				n = 3
			}
			ev.AddCounterAtETB("+1/+1", n)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Thunderous Velocipede: enters with additional +1/+1 counters",
	}
}
