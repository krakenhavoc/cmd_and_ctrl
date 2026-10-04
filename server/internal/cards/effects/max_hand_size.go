package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// max_hand_size.go — the card side of ADR 0113 §3 (#2074): statics
// that set or change a player's maximum hand size. The engine half, the
// CR 613.11 timestamp-order fold, is game/max_hand_size.go.
//
// Append-only, mechanic-named (the clone gate's convention): shared
// declarations for the maximum-hand-size cards live here.

// handSizeStatics is the CardDef's HandSize list: Spec.NoMaxHandSize
// (the #338 shorthand, with its designation gate) folded in front of
// Spec.HandSize. Nil when the card declares neither.
func handSizeStatics(spec Spec) []game.HandSizeStatic {
	if !spec.NoMaxHandSize {
		return spec.HandSize
	}
	out := make([]game.HandSizeStatic, 0, 1+len(spec.HandSize))
	out = append(out, game.HandSizeStatic{
		Players: game.HandSizeYou,
		Kind:    game.HandSizeNoMaximum,
		When:    spec.NoMaxHandSizeWhen,
	})
	return append(out, spec.HandSize...)
}

// checkHandSize refuses the two declarations no printed card makes: a
// maximum SET below zero (CR 107.1b clamps a result, but "is -1" is not
// a printed number) and a modification by zero, which does nothing and
// is a typo for a missing N. An entry whose number is read at the time
// (Dynamic) declares no N, so only its kind is checked.
func checkHandSize(spec Spec) {
	for i, h := range spec.HandSize {
		if h.Dynamic != nil {
			if h.Kind != game.HandSizeSet && h.Kind != game.HandSizeModify && h.Kind != game.HandSizeNoMaximum {
				panic(fmt.Sprintf("effects.Register: %q HandSize[%d] has unknown kind %d", spec.Name, i, h.Kind))
			}
			continue
		}
		switch h.Kind {
		case game.HandSizeSet:
			if h.N < 0 {
				panic(fmt.Sprintf("effects.Register: %q HandSize[%d] sets a maximum hand size of %d — a set value is 0 or more", spec.Name, i, h.N))
			}
		case game.HandSizeModify:
			if h.N == 0 {
				panic(fmt.Sprintf("effects.Register: %q HandSize[%d] modifies a maximum hand size by 0", spec.Name, i))
			}
		case game.HandSizeNoMaximum:
			if h.N != 0 {
				panic(fmt.Sprintf("effects.Register: %q HandSize[%d] is \"no maximum hand size\" with an N", spec.Name, i))
			}
		default:
			panic(fmt.Sprintf("effects.Register: %q HandSize[%d] has unknown kind %d", spec.Name, i, h.Kind))
		}
	}
}

// EachOpponentsMaxHandSizeReducedBy is "Each opponent's maximum hand
// size is reduced by n" (Jin-Gitaxias, Core Augur; Gnat Miser; Locust
// Miser).
func EachOpponentsMaxHandSizeReducedBy(n int) game.HandSizeStatic {
	return game.HandSizeStatic{Players: game.HandSizeEachOpponent, Kind: game.HandSizeModify, N: -n}
}

// YourMaxHandSizeIs is "Your maximum hand size is n" (Null Profusion).
func YourMaxHandSizeIs(n int) game.HandSizeStatic {
	return game.HandSizeStatic{Players: game.HandSizeYou, Kind: game.HandSizeSet, N: n}
}

// playACardHandOfTwo is the whole of Null Profusion and Recycle, which
// print the same three lines: "Skip your draw step. Whenever you play a
// card, draw a card. Your maximum hand size is two."
func playACardHandOfTwo(oracleID, name string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SkipYourDrawStep()},
		Triggered: []game.TriggeredAbility{
			WheneverYouPlayACard(name+" — draw a card", Do(DrawCards{N: 1})),
		},
		HandSize: []game.HandSizeStatic{YourMaxHandSizeIs(2)},
	}
}

// PlayersHaveNoMaxHandSize is "Players have no maximum hand size"
// (Price of Knowledge): every player, the controller included.
func PlayersHaveNoMaxHandSize() game.HandSizeStatic {
	return game.HandSizeStatic{Players: game.HandSizeEachPlayer, Kind: game.HandSizeNoMaximum}
}

// YourMaxHandSizeChangedBy is "Your maximum hand size is increased by n"
// (Minamo Scrollkeeper, Trusted Advisor) or, with a negative n,
// "reduced by" (the Thought Beasts).
func YourMaxHandSizeChangedBy(n int) game.HandSizeStatic {
	return game.HandSizeStatic{Players: game.HandSizeYou, Kind: game.HandSizeModify, N: n}
}

// AtYourEndStepDrawUpTo is "At the beginning of your end step, if you
// have fewer than n cards in hand, draw cards equal to the difference"
// (Doctor Octopus, Master Planner; The Ten Rings). An intervening-if
// (CR 603.4): checked as the step begins and again as the trigger
// resolves, the difference counted then.
func AtYourEndStepDrawUpTo(name string, n int) game.TriggeredAbility {
	label := fmt.Sprintf("%s — draw up to %d cards in hand", name, n)
	return On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller && b14HandSize(g, source.Controller) < n
	}, label, func(g *game.Game, item *game.StackItem) error {
		short := n - b14HandSize(g, item.Controller)
		if short <= 0 {
			return nil
		}
		return DrawCards{Player: item.Controller, N: short}.Apply(NewContext(g, item))
	})
}
