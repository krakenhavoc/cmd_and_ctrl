package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tapland_helpers.go — "This land enters tapped UNLESS <condition>",
// the shared body of the three conditional-dual cycles that sit near
// the top of the Commander play-rate list: the Innistrad checklands,
// the Battle for Zendikar battle lands and the Battlebond / Commander
// Legends bond lands. Twenty cards, one machinery shape.
//
// Its own file rather than helpers.go or enters_tapped.go, per the
// convention #231 set: concurrent card batches collide on shared
// helper files, and enters_tapped.go is exactly the kind of file two
// branches reach for at once.
//
// The condition is checked ONCE, as the land enters, and never again
// (CR 614.12 — a replacement effect looks at the game state just
// before the event). A checkland that entered tapped stays tapped
// when a Swamp shows up later, which is the printed behaviour.

// SelfEntersTappedUnless is the conditional cousin of
// SelfEntersTapped: the land enters tapped only when `untapped`
// reports false at entry time.
//
// The condition lives in AppliesTo rather than inside Replace so that
// a land meeting its condition contributes NO applicable replacement
// at all. That matters for CR 616: a Replace that decided to do
// nothing would still have been one of the effects the controller was
// asked to order, and the prompt would appear for a land that was
// always going to enter untapped.
//
// `untapped` receives the entering land's controller as ev.Actor
// rather than src.Controller. The two are the same for a land in
// play, but the battlefield-entry path stamps Card.Controller AFTER
// the replacement pipeline runs — so src.Controller is whatever the
// card carried in the previous zone, and the Actor is the only field
// guaranteed correct at this point.
func SelfEntersTappedUnless(untapped func(g *game.Game, controller uuid.UUID) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
				return false
			}
			if src == nil || ev.CardID != src.InstanceID {
				return false
			}
			return !untapped(g, ev.Actor)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
	}
}

// youControlLandTyped is the checkland condition — "unless you
// control an Island or a Swamp".
//
// The printed text names the land TYPE, not a basic land, so a
// Watery Grave, a Snow-Covered Island or another checkland already in
// play all satisfy Drowned Catacomb. IsLandWithSubtype is the same
// matcher the fetchlands use for the same reason.
//
// Needles MUST be lowercase — containsFoldASCII folds the haystack
// and not the needle, the trap already documented on
// eventCardHasType.
func youControlLandTyped(a, b string) func(*game.Game, uuid.UUID) bool {
	first, second := IsLandWithSubtype(a), IsLandWithSubtype(b)
	return func(g *game.Game, controller uuid.UUID) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller {
				continue
			}
			if first(c) || second(c) {
				return true
			}
		}
		return false
	}
}

// youControlTwoOrMoreBasics is the battle-land condition — "unless
// you control two or more basic lands". Counts basics only, so the
// dual cycle does not bootstrap itself off other nonbasics.
func youControlTwoOrMoreBasics(g *game.Game, controller uuid.UUID) bool {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && IsBasicLand(c) {
			n++
			if n >= 2 {
				return true
			}
		}
	}
	return false
}

// youHaveTwoOrMoreOpponents is the bond-land condition — "unless you
// have two or more opponents". True at any four-player table, which
// is what makes the cycle a Commander staple and a Constructed
// non-card; false in a two-player game and false once the table has
// been ground down to a single opponent.
//
// Eliminated seats are not opponents (CR 800.4a), so a bond land
// played after two players are dead enters tapped. That is the
// printed behaviour and it is why the count runs over live seats
// rather than over len(g.Seats).
func youHaveTwoOrMoreOpponents(g *game.Game, controller uuid.UUID) bool {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		n++
	}
	return n >= 2
}

// dualManaAbility lives in conditional_enters_tapped.go — same package,
// and the Hashaton land batch declared it first. Identical body.
