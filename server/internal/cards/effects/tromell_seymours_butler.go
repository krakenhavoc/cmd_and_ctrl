package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tromell, Seymour's Butler — Legendary Creature — Elf Advisor {2}{G},
// 2/3:
//
//	"Each other nontoken creature you control enters with an additional
//	 +1/+1 counter on it.
//	 {1}, {T}: Proliferate X times, where X is the number of nontoken
//	 creatures you control that entered this turn. (To proliferate,
//	 choose any number of permanents and/or players, then give each
//	 another counter of each kind already there.)"
//
// The counter is Renata's CR 614 entry replacement narrowed to nontoken
// creatures, so Hardened Scales and Doubling Season see it. X is counted
// as the ability resolves: nontoken creatures you control now that
// entered this turn as their current objects (Game.EnteredThisTurn),
// Tromell included if he did. Each proliferate is its own prompt and its
// own keyword action, so a proliferate replacement applies to each, and
// they are asked one after another.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f088817a-570d-4480-bc59-f2ad6675a71b",
		Name:         "Tromell, Seymour's Butler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{tromellExtraCounter()},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Proliferate X times, where X is the number of nontoken creatures you control that entered this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsCreature() && !c.IsToken() && g.EnteredThisTurn(c.InstanceID) {
						x++
					}
				}
				return proliferateTimes(g, item.Controller, item.SourceCardID, x)
			},
		}},
	})
}

// proliferateTimes proliferates n times in a row for `controller`, each
// one asked once the one before it is answered.
func proliferateTimes(g *game.Game, controller, source uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	return g.ProliferateChoosingForEffect(controller, source, func(g *game.Game) error {
		return proliferateTimes(g, controller, source, n-1)
	})
}

// tromellExtraCounter is "each other nontoken creature you control
// enters with an additional +1/+1 counter on it".
func tromellExtraCounter() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: otherNontokenCreatureYouControlEntering,
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Tromell, Seymour's Butler — each other nontoken creature you control enters with an additional +1/+1 counter",
	}
}

func otherNontokenCreatureYouControlEntering(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if src == nil || ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || ev.CardID == src.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && !c.IsToken() && c.Controller == src.Controller
}
