package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// conditional_enters_tapped.go — "this land enters tapped UNLESS
// <condition>", the one mechanic behind three whole land cycles
// (checklands, fastlands, slowlands) and the reason a modern
// two-colour mana base is mostly untapped.
//
// Its own file rather than enters_tapped.go so a concurrent batch
// touching that file doesn't collide, per the convention #231 set.
//
// The condition is evaluated PRE-PUSH, in the replacement pipeline,
// which is both correct and load-bearing:
//
//   - The entering land is not on the battlefield yet, so a walk of
//     g.Battlefield.Cards counts exactly the OTHER lands its
//     controller controls. That is literally what a fastland
//     ("two or fewer other lands") and a slowland ("two or more
//     other lands") ask for, with no special-casing.
//   - A replacement means the land was never untapped on the
//     battlefield. An OnETB tap would enter untapped and tap a beat
//     later, emitting EventTapCard, which anything watching for a
//     tap or for an untapped permanent entering would misread.
//
// This works because the pipeline consults an ENTERING card's own
// replacements as a third gathering block (added with the Temple
// cycle) — the ordinary catalog walk cannot see a card that is not
// on the battlefield yet.

// EntersTappedUnless is SelfEntersTapped gated on a condition read
// off the game state as the permanent enters. The permanent enters
// tapped when `cond` is FALSE — phrased to match the printed text,
// which always states the condition under which the drawback does
// NOT apply.
//
// `src` is the entering card itself; read src.Controller for "you"
// and treat the battlefield as containing only OTHER permanents.
// A nil cond degrades to an unconditional SelfEntersTapped.
func EntersTappedUnless(cond func(g *game.Game, src *game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
				return false
			}
			if src == nil || ev.CardID != src.InstanceID {
				return false
			}
			return cond == nil || !cond(g, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
	}
}

// countOtherLandsControlled counts the lands src's controller
// controls, not counting src. The exclusion is belt-and-braces: the
// entering card is not on the battlefield when this runs, so the
// walk would not find it anyway — but the same helper is useful from
// a battlefield-side caller, and a silently off-by-one land count is
// exactly the sort of thing that is invisible until a fastland
// enters tapped on turn three.
func countOtherLandsControlled(g *game.Game, src *game.Card) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == src.InstanceID {
			continue
		}
		if c.Controller == src.Controller && c.IsLand() {
			n++
		}
	}
	return n
}

// otherLandsAtMost is a fastland's condition: "unless you control
// two or fewer other lands".
func otherLandsAtMost(n int) func(*game.Game, *game.Card) bool {
	return func(g *game.Game, src *game.Card) bool {
		return countOtherLandsControlled(g, src) <= n
	}
}

// otherLandsAtLeast is a slowland's condition: "unless you control
// two or more other lands".
func otherLandsAtLeast(n int) func(*game.Game, *game.Card) bool {
	return func(g *game.Game, src *game.Card) bool {
		return countOtherLandsControlled(g, src) >= n
	}
}

// otherLandsWithSubtypeAtLeast is the sanctuary-land condition:
// "unless you control three or more other Islands". The count reads
// EFFECTIVE subtypes (game.Card.HasSubtype), so a land another
// layer-4 effect made an Island counts and a Blood Moon-ed nonbasic
// does not — the same rule the intrinsic mana ability follows
// (CR 305.6).
//
// `subtype` is matched case-insensitively by HasSubtype, so pass it
// the way the rest of the catalog does ("island").
func otherLandsWithSubtypeAtLeast(subtype string, n int) func(*game.Game, *game.Card) bool {
	return func(g *game.Game, src *game.Card) bool {
		count := 0
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == src.InstanceID || c.Controller != src.Controller {
				continue
			}
			if c.IsLand() && c.HasSubtype(subtype) {
				count++
			}
		}
		return count >= n
	}
}

// dualManaAbility is the pipe-syntax "{T}: Add {X} or {Y}" every
// two-colour land in this batch carries. Two separate one-colour
// abilities would also work but would clutter the activation menu
// with a fixed pair.
func dualManaAbility(a, b string) ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{" + a + "|" + b + "}",
		Label:    "Add {" + a + "} or {" + b + "}",
	}
}
