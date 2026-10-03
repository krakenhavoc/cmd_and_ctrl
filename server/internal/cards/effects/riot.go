package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// riot.go — the catalog half of riot (CR 702.136) and unleash (CR
// 702.98), ADR 0109 §10 (#1556).
//
// Both keywords are engine tokens (game.KeywordRiot, game.KeywordUnleash)
// whose entry replacements the engine derives from the CR 614.12
// look-ahead (game/riot.go). A card that prints one declares it in
// PrintedKeywords and writes nothing else; a card that GRANTS one is an
// ordinary layer-6 keyword grant, and a creature entering under it is
// asked as it enters. This file holds the grant shapes more than one
// riot or unleash card prints, so no card file grows its own copy.

// NontokenCreaturesYouControlHave is "Nontoken creatures you control
// have <keyword>" — Rhythm of the Wild's and Uncivil Unrest's riot. The
// controller is the permanent's would-be controller as it enters, which
// the look-ahead stamps, so an opponent's grant gives your creature
// nothing.
func NontokenCreaturesYouControlHave(keyword string) game.StaticAbility {
	return KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && !target.IsToken() && target.Controller == source.Controller
	}, keyword)
}

// CreaturesYouControlWithCountersHave is "<other> creatures you control
// with <a counter / a +1/+1 counter> on them have <keyword>" — Tesak's
// haste for any counter (kind "") and Exava's for "each other creature
// you control with a +1/+1 counter on it". Counters are read off the
// live card: a counter is a marker the layer pass reads, and placing or
// removing one bumps the layer version.
func CreaturesYouControlWithCountersHave(kind string, others bool, keyword string) game.StaticAbility {
	return KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		if !target.IsCreature() || target.Controller != source.Controller {
			return false
		}
		if others && target.InstanceID == source.InstanceID {
			return false
		}
		return hasCounterOfKind(*target, kind)
	}, keyword)
}

// ThisHasWhileItHasCounter is "This creature has <keyword> as long as it
// has a <kind> counter on it" — Chaos Imps' trample.
func ThisHasWhileItHasCounter(kind, keyword string) game.StaticAbility {
	return KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && hasCounterOfKind(*target, kind)
	}, keyword)
}

// hasCounterOfKind reports whether the card carries a counter of `kind`,
// or of any kind when `kind` is empty.
func hasCounterOfKind(c game.Card, kind string) bool {
	if kind != "" {
		return c.Counters[kind] > 0
	}
	for _, n := range c.Counters {
		if n > 0 {
			return true
		}
	}
	return false
}
