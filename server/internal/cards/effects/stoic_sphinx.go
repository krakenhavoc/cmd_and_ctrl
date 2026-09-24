package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stoic Sphinx — Creature — Sphinx {2}{U}{U}, 5/3:
//
//	"Flash
//	 Flying
//	 This creature has hexproof as long as you haven't cast a spell
//	 this turn."
//
// A defensive flash blocker: it comes down at instant speed, holds
// off a removal spell for as long as its controller keeps priority
// closed, and loses that protection the instant they cast anything —
// including the spell that answers whatever just threatened it.
//
// # The static is the seam this card proves (#1325)
//
// "As long as you haven't cast a spell this turn" is a Layer 6
// keyword grant whose AppliesTo reads Game.CastTallyFor — a per-turn
// count that is not on the battlefield, not a counter and not tap
// state, so nothing invalidated the layer cache when it changed. Hand
// size (#74), life total (#1117) and attacking status (#1218) were
// the earlier non-battlefield inputs a static could read; the
// spells-cast count is a fifth quarter on the same "Layer invalidation
// on hand / life / attack / graveyard state" row (docs/engine-seams.md)
// that #1325 adds and closes the same way: StaticAbility.DependsOnSpellsCast
// declares the dependency, and layer_listener.go's EventCast arm drops
// the cache the instant the tally the condition reads is bumped.
// Without the fix the Sphinx would keep hexproof through the cast that
// should have turned it off, until some unrelated event happened to
// invalidate.
//
// See internal/game/layer_invalidation_cast_test.go for the
// engine-level proof and this file's test for the card's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d924c13e-f266-47b0-9711-ba9a9b9df1cb",
		Name:            "Stoic Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static:          []game.StaticAbility{stoicSphinxHexproofWhileNoSpellCastYet()},
	})
}

// stoicSphinxHexproofWhileNoSpellCastYet is the card's own clause,
// pulled out to its own function only so the doc comment above can
// point at one name rather than an anonymous literal.
func stoicSphinxHexproofWhileNoSpellCastYet() game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && g.CastTallyFor(source.Controller).Total == 0
	}, "hexproof")
	ab.DependsOnSpellsCast = true
	return ab
}
