package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divide_amounts.go — #1657, ADR 0065's 2026-09-28 (b) amendment: the
// registered rules a divided clause's amount can name with DivideBy.
// Each is a package-level var so it is registered exactly once, at
// init, under a key a restored game finds again; see
// game.RegisterDivideAmount for why the clause carries a key and not a
// func.
//
// Every rule is read ONCE, at announce — a cast, an activation, or the
// moment a trigger's target walk opens (CR 601.2d / 602.2b / 603.3d) —
// and the number it answers is the number divided, whatever happens to
// the board afterwards.

// DivideLandsYouControl is "X damage divided as you choose …, where X
// is the number of lands you control" (Ureni, the Song Unending).
// "You" is the announcing player. Ureni's ruling: "The value of X
// won't change even if the number of lands you control changes after
// that point."
var DivideLandsYouControl = game.RegisterDivideAmount("lands-you-control",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: b02CountLandsControlledBy(g, a.Controller)}
	})

// DivideSourcePower is "damage equal to its power divided as you
// choose" on a trigger whose source may be gone by the time the
// ability is put on the stack (Orca, Siege Demon's dies trigger): the
// power the source last had on the battlefield, counters included
// (CR 603.10, b13LastKnownPower). A source with no last-known state —
// a cast or activation reaching this rule — reads its live power.
var DivideSourcePower = game.RegisterDivideAmount("source-power",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		if a.SourceLKI != nil {
			return game.DivideSpec{Total: b13LastKnownPower(g, a.Source, *a.SourceLKI)}
		}
		c, ok := g.LookupCardForEffect(a.Source)
		if !ok {
			return game.DivideSpec{}
		}
		return game.DivideSpec{Total: c.CurrentPower()}
	})

// DivideLifeYouGainedThisTurn is "up to that many", where that many is
// the life you gained this turn (Lathiel, the Bounteous Dawn). The
// turn's total, not the net change of the life total — Lathiel's
// ruling — which is TurnTally's LifeGained cell. "Gaining more life in
// response to the triggered ability won't change how many +1/+1
// counters will be distributed."
var DivideLifeYouGainedThisTurn = game.RegisterDivideAmount("life-you-gained-this-turn",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: b15LifeGainedThisTurn(g, a.Controller)}
	})

// DivideTwoOrXIfMadness is Avacyn's Judgment's amount: "2 damage
// divided as you choose … If this spell's madness cost was paid, it
// deals X damage divided as you choose among those permanents and/or
// players instead." The madness cost is {X}{R}, so the X is the one
// announced for that cast; a cast that paid the printed {1}{R} has no
// X at all and divides 2. The claimed alternative cost is the one
// StackItem.AltCost records (CR 702.35b).
var DivideTwoOrXIfMadness = game.RegisterDivideAmount("two-or-x-if-madness",
	func(_ *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		if a.AltCost == game.AltCostKeyMadness {
			return game.DivideSpec{FromX: true}
		}
		return game.DivideSpec{Total: 2}
	})
