package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rockalanche — Sorcery — Lesson {2}{G}:
//
//	"Earthbend X, where X is the number of Forests you control.
//	 (Target land you control becomes a 0/0 creature with haste that's
//	 still a land. Put X +1/+1 counters on it. When it dies or is
//	 exiled, return it to the battlefield tapped.)
//	 Flashback {5}{G} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The second earthbend proof, and it is here for the two things
// Earthbending Lesson cannot show: a COUNT read at resolution, and the
// keyword composing with an existing one.
//
// # X IS COUNTED AT RESOLUTION, AND IT CAN BE ZERO
//
// Not an announced X — nothing about this spell's cost varies — so the
// count is read when the spell resolves, over the Forests the
// controller has THEN. A Forest that died in response is not counted,
// and one that entered is.
//
// A mono-green deck casting this on turn four counts three or four; a
// deck with no Forest at all counts nothing, and "earthbend 0" is a
// real instruction rather than a fizzle: the land becomes a 0/0
// creature with haste, gets no counters, dies to the toughness
// state-based action the next time anybody would get priority, and
// comes back tapped. That is the printed card, and it is the reason
// `KeywordAction.actsAtZeroCount` exists.
//
// "FORESTS" IS THE SUBTYPE, not "green lands" and not "basic Forests":
// a Stomping Ground, a Bayou and a Forest all count, and so does a
// land that a layer-4 effect has made a Forest, because `HasSubtype`
// reads the effective type line (#255). Snow-Covered Forest counts too.
// The predicate does NOT require the land to be untapped or to be
// anything but a Forest you control.
//
// # FLASHBACK IS THE OTHER HALF OF THE CARD
//
// Two earthbends over a game for seven mana total, which is what makes
// this a Lesson worth a slot: the first copy makes a 3/3-ish attacker,
// the flashback copy makes a second one out of a land that is doing
// nothing in the late game. The zone and the price are declared
// separately for the reason Artful Dodge's comment gives — the
// registry panics on a zone-bound cost whose zone was not listed.
//
// The two copies earthbend two DIFFERENT lands only if you target
// differently; earthbending the same land twice is legal, adds the
// second batch of counters, and does not stack a second return (see
// `scheduleEarthbendReturnLocked`).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d1ecfad3-e79a-447b-a932-68ef728eaf1f",
		Name:             "Rockalanche",
		Completeness:     CompletenessFull,
		Targets:          EarthbendTargets(),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{G}")},
		OnResolve:        EarthbendFirstTarget(rockalancheForests),
	})
}

// rockalancheForests is "the number of Forests you control", counted
// at resolution over the effective type line.
func rockalancheForests(ctx *Context) int {
	return countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
		return c.HasSubtype("Forest")
	})
}
